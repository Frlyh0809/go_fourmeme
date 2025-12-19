package v3

import (
	"context"
	"fmt"
	"go_fourmeme/client"
	"go_fourmeme/trade"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"go_fourmeme/config"
	configentity "go_fourmeme/entity/config"
	"go_fourmeme/log"
	"go_fourmeme/manager"
	"go_fourmeme/utils"
)

// BuyTokenViaManager 一级市场买入 (通过 Helper3 buyWithEth)
func BuyTokenViaManager(target *configentity.MonitorTarget, tokenAddr string) (string, error) {
	ethClient := manager.GetEthClient()
	info := client.GetTokenStatus(tokenAddr)
	if info.Status != client.StatusTrading {
		log.LogWarn("Token 状态非 TRADING (%s)，跳过买入: %s", info.Status, tokenAddr[:10])
		return "", nil
	}

	amountInWei := trade.ToWei(target.BuyAmountBNB)
	if amountInWei.Sign() == 0 {
		return "", fmt.Errorf("买入金额为0")
	}

	log.LogInfo("准备买入 Token: %s | AmountInWei: %s | AmountBNB: %s", tokenAddr, amountInWei.String(), target.BuyAmountBNB.String())

	minAmountOut, err := client.LocalCalcMinAmountOut(info, amountInWei, target.SlippageTolerance)
	if err != nil {
		log.LogWarn("本地滑点计算失败，使用保守值: %v1", err)
		// 保守 fallback：amountInWei * (1 - slippage)，假设slippage是0-1的小数
		slippagePercent := new(big.Int).SetInt64(int64(target.SlippageTolerance * 100))
		minAmountOut = new(big.Int).Mul(amountInWei, new(big.Int).Sub(big.NewInt(100), slippagePercent))
		minAmountOut.Div(minAmountOut, big.NewInt(100))
	}

	// 恢复滑点预估 (假设 calcSlippageMinOut 已定义；如果未定义，可注释)
	// minAmountOut, err = calcSlippageMinOut(ethClient, tokenAddr, amountInWei, target.SlippageTolerance)
	// if err != nil {
	// 	return "", fmt.Errorf("滑点计算失败: %v1", err)
	// }

	helperABI := utils.GetABI("TokenManagerHelper")
	if helperABI == nil {
		return "", fmt.Errorf("Helper ABI 未加载")
	}
	balance, balanceErr := ethClient.PendingBalanceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	if balanceErr != nil {
		return "", fmt.Errorf("PendingBalanceAt 调用失败: %v1", balanceErr)
	}
	log.LogInfo("-- wallet:%s balance:%d", config.BSCChain.WalletAddress, balance)

	// 先用 tryBuy 预模拟，确认quote和预计输出
	tryBuyInput, err := helperABI.Pack("tryBuy", common.HexToAddress(tokenAddr), big.NewInt(0), amountInWei) // amount=0: 按funds买
	if err != nil {
		return "", fmt.Errorf("Pack tryBuy 失败: %v1", err)
	}
	helperAddr := common.HexToAddress(config.TokenManagerHelper3)
	callMsg := ethereum.CallMsg{
		From:  common.HexToAddress(config.BSCChain.WalletAddress),
		To:    &helperAddr,
		Value: amountInWei,
		Data:  tryBuyInput,
	}
	result, err := ethClient.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return "", fmt.Errorf("tryBuy 调用失败: %v1", err)
	}
	// Unpack tryBuy 返回值 (假设返回: uint256 amountMsgValue, uint256 amountApproval, uint256 estimatedAmount, address approvalToken)
	unpacked, err := helperABI.Unpack("tryBuy", result)
	if err != nil {
		return "", fmt.Errorf("Unpack tryBuy 失败: %v1", err)
	}
	amountMsgValue := unpacked[0].(*big.Int)
	amountApproval := unpacked[1].(*big.Int)
	estimatedAmount := unpacked[2].(*big.Int)
	approvalToken := unpacked[3].(common.Address)

	log.LogInfo("tryBuy 结果: MsgValue=%s, Approval=%s, EstimatedOut=%s, ApprovalToken=%s", amountMsgValue.String(), amountApproval.String(), estimatedAmount.String(), approvalToken.Hex())

	// 检查是否为BNB quote
	if amountMsgValue.Cmp(amountInWei) != 0 || amountApproval.Sign() > 0 {
		return "", fmt.Errorf("非纯BNB盘子 (需Approval: %s for %s)，请用buyWithToken或其他Helper", amountApproval.String(), approvalToken.Hex())
	}

	// 更新minAmountOut为estimatedAmount * (1 - slippage)，更精确
	minAmountOut = new(big.Int).Mul(estimatedAmount, new(big.Int).Sub(big.NewInt(100), new(big.Int).SetInt64(int64(target.SlippageTolerance*100))))
	minAmountOut.Div(minAmountOut, big.NewInt(100))

	// Pack buyWithEth: origin=0, token, to=wallet, funds=amountInWei, minAmount
	input, err := helperABI.Pack("buyWithEth",
		big.NewInt(0), // origin (requestId, 0 为默认)
		common.HexToAddress(tokenAddr),
		common.HexToAddress(config.BSCChain.WalletAddress), // to (接收者)
		amountInWei,
		minAmountOut,
	)
	if err != nil {
		return "", fmt.Errorf("Pack buyWithEth 失败: %v1", err)
	}

	// 构建交易
	fmt.Println("config.BSCChain.WalletAddress : ", config.BSCChain.WalletAddress)
	nonce, nonceErr := ethClient.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	if nonceErr != nil {
		return "", fmt.Errorf("获取 nonce 失败: %v1", nonceErr)
	}
	log.LogInfo("当前钱包 nonce: %d", nonce)

	gasPrice, gasErr := ethClient.SuggestGasPrice(context.Background())
	if gasErr != nil {
		log.LogWarn("获取 gasPrice 失败，使用默认1 gwei: %v1", gasErr)
		gasPrice = big.NewInt(1000000000) // 1 gwei 默认，可调整
	} else {
		// 稍提高gasPrice防pending
		gasPrice.Mul(gasPrice, big.NewInt(11)).Div(gasPrice, big.NewInt(10))
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &helperAddr,
		Value:    amountInWei, // payable BNB
		Gas:      1200000,     // 提高gas limit
		GasPrice: gasPrice,
		Data:     input,
	})

	privateKey, err := trade.GetPrivateKey()
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(config.BSCChain.ChainID), privateKey)
	if err != nil {
		return "", err
	}

	err = ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		//database.SaveTxRecord(&po.TransactionRecord{
		//	Type:      "buy_primary",
		//	TokenAddr: tokenAddr,
		//	AmountIn:  amountInWei,
		//	Status:    "failed",
		//	ErrorMsg:  err.Error(),
		//	Timestamp: time.Now(),
		//})
		return "", fmt.Errorf("发送交易失败: %v1", err)
	}

	txHash := signedTx.Hash().Hex()
	log.LogInfo("一级买入提交成功 Tx: %s Token: %s Amount: %s BNB", txHash, tokenAddr[:10], target.BuyAmountBNB.Text('f', 6))

	// 等待收据 + 添加持仓
	receipt, err := trade.WaitForReceipt(ethClient, signedTx.Hash())
	if err != nil {
		log.LogWarn("收据获取失败，使用估算: %v1", err)
		trade.AddPositionFromEstimate(tokenAddr, txHash, target, minAmountOut)
		return txHash, nil
	}

	actualOut := trade.ExtractTokenOutFromReceipt(receipt, tokenAddr)
	if actualOut == nil {
		actualOut = minAmountOut
	}
	trade.AddPositionFromReceipt(tokenAddr, txHash, target, actualOut)

	return txHash, nil
}

// SellTokenViaManager 一级市场卖出 (通过 Helper3 sellForEth)
func SellTokenViaManager(tokenAddr string, amountIn *big.Int, slippage float64) (string, error) {
	client := manager.GetEthClient()

	if amountIn.Sign() == 0 {
		return "", fmt.Errorf("卖出金额为0")
	}

	// 滑点预估 (calcEthOut)
	minFunds, err := calcSlippageMinOutForSell(client, tokenAddr, amountIn, slippage)
	if err != nil {
		return "", err
	}

	helperABI := utils.GetABI("TokenManagerHelper")
	input, err := helperABI.Pack("sellForEth",
		big.NewInt(0), // origin
		common.HexToAddress(tokenAddr),
		amountIn,
		minFunds,
	)
	if err != nil {
		return "", err
	}

	// 构建交易 (非 payable)
	nonce, _ := client.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	gasPrice, _ := client.SuggestGasPrice(context.Background())
	helperAddr := common.HexToAddress(config.TokenManagerHelper3)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &helperAddr,
		Value:    big.NewInt(0),
		Gas:      800000,
		GasPrice: gasPrice,
		Data:     input,
	})

	privateKey, err := trade.GetPrivateKey()
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(config.BSCChain.ChainID), privateKey)
	if err != nil {
		return "", err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	txHash := signedTx.Hash().Hex()
	log.LogInfo("一级卖出提交成功 Tx: %s Token: %s Amount: %s", txHash, tokenAddr[:10], amountIn.String())

	return txHash, nil
}

// calcSlippageMinOutForSell 卖出滑点预估 (calcEthOut)
func calcSlippageMinOutForSell(client *ethclient.Client, tokenAddr string, amountIn *big.Int, slippage float64) (*big.Int, error) {
	helperABI := utils.GetABI("TokenManagerHelper")
	if helperABI == nil {
		return new(big.Int).Mul(amountIn, big.NewInt(int64(100-int(slippage*100)))), fmt.Errorf("Helper ABI 未加载")
	}

	input, err := helperABI.Pack("calcEthOut", common.HexToAddress(tokenAddr), amountIn)
	if err != nil {
		return nil, err
	}
	tokenManagerHelper3 := common.HexToAddress(config.TokenManagerHelper3)
	callMsg := ethereum.CallMsg{
		To:   &tokenManagerHelper3,
		Data: input,
	}

	output, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return nil, err
	}

	var expectedOut *big.Int
	err = helperABI.UnpackIntoInterface(&expectedOut, "calcEthOut", output)
	if err != nil {
		return nil, err
	}

	minOut := new(big.Int).Mul(expectedOut, big.NewInt(int64(100-int(slippage*100))))
	minOut.Div(minOut, big.NewInt(100))
	return minOut, nil
}
