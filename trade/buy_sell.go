// trade/buy_sell.go
package trade

import (
	"context"
	"fmt"
	"go_fourmeme/entity/po"
	"go_fourmeme/manager"
	"math/big"
	"time"

	"go_fourmeme/config"
	"go_fourmeme/database"
	"go_fourmeme/entity"
	"go_fourmeme/log"
	"go_fourmeme/utils"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// BuyTokenViaManager 一级市场买入（通过 Fourmeme TokenManager2）
// 成功后会自动记录持仓到全局 positions（main 中监控盈亏）
func BuyTokenViaManager(client *ethclient.Client, tokenAddr string, amountInBNB *big.Int, slippage float64, target *config.MonitorTarget) (string, error) {
	if utils.GetABI("TokenManager2") == nil {
		err := fmt.Errorf("TokenManager2 ABI 未加载")
		log.LogError("%v", err)
		return "", err
	}

	// --------------------- 1. 计算 minAmountOut（滑点保护） ---------------------
	// 注意：一级市场是 bonding curve，通常没有公开 reserves，可通过模拟调用或预估
	// 这里先简单使用 slippage 直接计算一个保守的 minOut（实际项目可调用 curve 查询函数）
	// 假设预期输出 ≈ amountInBNB * 当前价格因子（可从事件估算或固定比例）
	minAmountOut := new(big.Int).Mul(amountInBNB, big.NewInt(90)) // 示例：90% 作为最小输出（根据实际调整）
	minAmountOut = new(big.Int).Div(minAmountOut, big.NewInt(100))

	// --------------------- 2. 使用 ABI pack 调用 buy 方法 ---------------------
	// 请根据 TokenManager2.lite.abi 中的实际方法名替换 "buy" 或 "buyToken"
	// 常见可能方法名：buy, buyToken, buyTokenAMAP, depositAndBuy 等
	input, err := utils.GetABI("TokenManager2").Pack(
		"buyTokenAMAP", // <--- 请务必替换为 ABI 中正确的函数名 !!!
		common.HexToAddress(tokenAddr),
		big.NewInt(0),
		minAmountOut, // 滑点保护参数（如果方法有）
	)
	if err != nil {
		// 如果 "buy" 不存在，可尝试其他常见方法名
		log.LogError("Pack buy 方法失败，尝试其他方法名: %v", err)
		return "", err
	}

	// --------------------- 3. 构建并签名交易 ---------------------
	nonce, err := client.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	if err != nil {
		return "", err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		gasPrice = big.NewInt(5000000000) // 5 Gwei 兜底
	}
	// 可根据需要上浮 gasPrice
	gasPrice.Mul(gasPrice, big.NewInt(int64(config.BSCChain.GasPriceMultiplier*100)))
	gasPrice.Div(gasPrice, big.NewInt(100))

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &config.DefaultFourmemeManagerAddr,
		Value:    amountInBNB, // payable，投入的 BNB
		Gas:      600000,      // 根据实际调用调整
		GasPrice: gasPrice,
		Data:     input,
	})

	privateKey, err := config.GetPrivateKey()
	if err != nil {
		return "", err
	}

	signer := types.NewEIP155Signer(config.BSCChain.ChainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return "", err
	}

	// --------------------- 4. 发送交易 ---------------------
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		// 记录失败
		database.SaveTxRecord(&po.TransactionRecord{
			Type:      "buy_manager",
			TokenAddr: tokenAddr,
			AmountIn:  amountInBNB,
			Slippage:  slippage,
			Status:    "failed",
			ErrorMsg:  err.Error(),
			Timestamp: time.Now(),
		})
		log.LogErrorMsg("一级市场买入发送失败 Token=%s err:%s", tokenAddr, err)
		return "", err
	}

	txHash := signedTx.Hash().Hex()
	log.LogInfo("一级市场买入提交成功 Tx=%s Token=%s Amount=%s BNB", txHash, tokenAddr, amountInBNB.String())

	// --------------------- 5. 等待收据并解析实际买入的 token 数量 ---------------------
	// 为了准确记录持仓，需要从收据的日志中提取 Transfer (Manager -> Wallet) 的数量
	receipt, err := waitForReceipt(client, signedTx.Hash())
	if err != nil {
		log.LogError("获取交易收据失败，将使用估算数量记录持仓: %v", err)
		// 仍记录一个估算持仓
		addPositionWithEstimate(tokenAddr, txHash, amountInBNB, minAmountOut, slippage, target)
		return txHash, nil
	}

	buyTokenAmount := extractBuyTokenAmountFromReceipt(receipt, tokenAddr)
	if buyTokenAmount == nil {
		buyTokenAmount = minAmountOut // fallback
	}

	// --------------------- 6. 计算平均价格并记录持仓 ---------------------
	buyAmountBNBFloat := new(big.Float).SetInt(amountInBNB)
	buyTokenAmountFloat := new(big.Float).SetInt(buyTokenAmount)
	buyPriceAvg := new(big.Float).Quo(buyAmountBNBFloat, buyTokenAmountFloat) // BNB per token

	addPosition(&entity.Position{
		TokenAddr:        tokenAddr,
		BuyTxHash:        txHash,
		BuyAmountBNB:     buyAmountBNBFloat,
		BuyTokenAmount:   buyTokenAmount,
		BuyPriceAvg:      buyPriceAvg,
		BuyTime:          time.Now(),
		TargetProfitMult: target.TakeProfitMultiple,
		TargetLossMult:   target.StopLossMultiple,
		Sold:             false,
	})

	// --------------------- 7. 成功记录数据库 ---------------------
	database.SaveTxRecord(&po.TransactionRecord{
		Type:      "buy_manager",
		TxHash:    txHash,
		TokenAddr: tokenAddr,
		AmountIn:  amountInBNB,
		AmountOut: buyTokenAmount,
		Slippage:  slippage,
		Status:    "success",
		Timestamp: time.Now(),
	})

	return txHash, nil
}

// waitForReceipt 等待交易上链
func waitForReceipt(client *ethclient.Client, hash common.Hash) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	for {
		receipt, err := client.TransactionReceipt(ctx, hash)
		if err == nil {
			return receipt, nil
		}
		if err == ethereum.NotFound {
			time.Sleep(2 * time.Second)
			continue
		}
		return nil, err
	}
}

// extractBuyTokenAmountFromReceipt 从收据中提取 Manager -> Wallet 的 Transfer 数量
func extractBuyTokenAmountFromReceipt(receipt *types.Receipt, tokenAddr string) *big.Int {
	transferTopic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	managerAddr := config.DefaultFourmemeManagerAddr
	walletAddr := common.HexToAddress(config.BSCChain.WalletAddress)

	for _, log := range receipt.Logs {
		if len(log.Topics) < 3 || log.Topics[0] != transferTopic {
			continue
		}
		if log.Address.Hex() != tokenAddr {
			continue
		}
		from := common.BytesToAddress(log.Topics[1].Bytes())
		to := common.BytesToAddress(log.Topics[2].Bytes())
		if from == managerAddr && to == walletAddr {
			return new(big.Int).SetBytes(log.Data)
		}
	}
	return nil
}

// addPositionWithEstimate 收据失败时的估算持仓
func addPositionWithEstimate(tokenAddr, txHash string, amountInBNB, estimatedOut *big.Int, slippage float64, target *config.MonitorTarget) {
	buyAmountBNBFloat := new(big.Float).SetInt(amountInBNB)
	buyTokenAmountFloat := new(big.Float).SetInt(estimatedOut)
	buyPriceAvg := new(big.Float).Quo(buyAmountBNBFloat, buyTokenAmountFloat)

	addPosition(&entity.Position{
		TokenAddr:        tokenAddr,
		BuyTxHash:        txHash,
		BuyAmountBNB:     buyAmountBNBFloat,
		BuyTokenAmount:   estimatedOut,
		BuyPriceAvg:      buyPriceAvg,
		BuyTime:          time.Now(),
		TargetProfitMult: target.TakeProfitMultiple,
		TargetLossMult:   target.StopLossMultiple,
		Sold:             false,
	})
}

// addPosition 实际添加持仓到全局 map（main 中使用）
func addPosition(pos *entity.Position) {
	manager.AddPosition(pos)
}

// BuyTokenSecondary: 二级市场（PancakeSwap）
func BuyTokenSecondary(client *ethclient.Client, tokenAddr string, amountIn *big.Int, slippage float64) (string, error) {
	routerABI := utils.GetABI("PancakeRouter")
	minOut := calculateMinOutFromReserves(client, amountIn, tokenAddr, slippage)

	path := []common.Address{common.HexToAddress(config.WBNB), common.HexToAddress(tokenAddr)}
	deadline := big.NewInt(time.Now().Add(20 * time.Minute).Unix())

	input, err := routerABI.Pack("swapExactETHForTokens", minOut, path, common.HexToAddress(config.BSCChain.WalletAddress), deadline)
	if err != nil {
		return "", err
	}

	// 类似LegacyTx构建，value=amountIn，to=Router
	// ... (代码类似BuyTokenViaManager)
	return txHash, nil
}

// calculateMinOut / calculateMinOutFromReserves: 滑点计算（查询pair reserves）
func calculateMinOutFromReserves(client *ethclient.Client, amountIn *big.Int, tokenAddr string, slippage float64) *big.Int {
	// 查询Pair reserves，使用Pair ABI getReserves
	// 实现略，返回minAmountOut = expected * (1 - slippage)
	return new(big.Int)
}

// privateKeyFromConfig: 从config获取ec.PrivateKey
func privateKeyFromConfig() *ecdsa.PrivateKey {
	// 实现
}
