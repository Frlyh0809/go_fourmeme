package primary_market

import (
	"context"
	"fmt"
	"go_fourmeme/client"
	"go_fourmeme/config"
	"go_fourmeme/database"
	configentity "go_fourmeme/entity/config"
	"go_fourmeme/entity/po"
	"go_fourmeme/log"
	"go_fourmeme/manager"
	"go_fourmeme/trade"
	"go_fourmeme/utils"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// BuyTokenViaManager2 一级市场买入 (纯 TokenManager2 V2 版本, BNB quote盘子)
func BuyTokenViaManager2(target *configentity.MonitorTarget, tokenAddr string) (string, error) {
	ethClient := manager.GetEthClient()
	info := client.GetTokenStatus(tokenAddr)
	if info.Status != client.StatusTrading {
		log.LogWarn("Token 状态非 TRADING (%s)，跳过买入: %s", info.Status, tokenAddr[:10])
		return "", nil
	}

	amountInWei := trade.ToWei(target.BuyAmountBNB) // 这就是 funds，通过 Value 传递
	if amountInWei.Sign() == 0 {
		return "", fmt.Errorf("买入金额为0")
	}

	log.LogInfo("准备买入 Token: %s | FundsWei: %s | AmountBNB: %s", tokenAddr, amountInWei.String(), target.BuyAmountBNB.String())

	// 本地计算 minAmount (滑点保护)
	minAmountOut, err := client.LocalCalcMinAmountOut(info, amountInWei, target.SlippageTolerance)
	if err != nil {
		log.LogWarn("本地滑点计算失败，使用保守值: %v", err)
		slippagePercent := big.NewInt(int64(target.SlippageTolerance * 100))
		minAmountOut = new(big.Int).Mul(amountInWei, new(big.Int).Sub(big.NewInt(100), slippagePercent))
		minAmountOut.Div(minAmountOut, big.NewInt(100))
	}
	if minAmountOut.Sign() == 0 {
		minAmountOut = big.NewInt(1) // 至少1，避免合约revert
	}

	managerABI := utils.GetABI("TokenManager2")
	if managerABI == nil {
		return "", fmt.Errorf("TokenManager2 ABI 未加载")
	}

	// Pack buyTokenAMAP: token, to, minAmount   (funds 通过 Value 传递，不在参数里)
	input, err := managerABI.Pack("buyTokenAMAP",
		big.NewInt(0),
		common.HexToAddress(tokenAddr),
		common.HexToAddress(config.BSCChain.WalletAddress), // to
		amountInWei,
		minAmountOut, // minAmount
	)
	if err != nil {
		return "", fmt.Errorf("Pack buyTokenAMAP 失败: %v", err)
	}

	// 构建交易
	nonce, nonceErr := ethClient.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	if nonceErr != nil {
		return "", fmt.Errorf("获取 nonce 失败: %v", nonceErr)
	}
	log.LogInfo("当前钱包 nonce: %d", nonce)

	gasPrice, gasErr := ethClient.SuggestGasPrice(context.Background())
	if gasErr != nil {
		log.LogWarn("获取 gasPrice 失败，使用默认1 gwei: %v", gasErr)
		gasPrice = big.NewInt(1000000000)
	} else {
		gasPrice.Mul(gasPrice, big.NewInt(11)).Div(gasPrice, big.NewInt(10)) // +10%
	}

	managerAddr := common.HexToAddress(config.TokenManager2) // 通常是 0x5c952063c7fc8610FFDB798152D69F0B9550762b

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &managerAddr,
		Value:    amountInWei, // payable funds (BNB金额)
		Gas:      1200000,
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
		database.SaveTxRecord(&po.TransactionRecord{
			Type:      "buy_primary_v2",
			TokenAddr: tokenAddr,
			AmountIn:  amountInWei,
			Status:    "failed",
			ErrorMsg:  err.Error(),
			Timestamp: time.Now(),
		})
		return "", fmt.Errorf("发送交易失败: %v", err)
	}

	txHash := signedTx.Hash().Hex()
	log.LogInfo("一级买入提交成功 (V2) Tx: %s Token: %s Amount: %s BNB", txHash, tokenAddr[:10], target.BuyAmountBNB.Text('f', 6))

	// 等待收据 + 添加持仓（保持你原来的逻辑）
	receipt, err := trade.WaitForReceipt(ethClient, signedTx.Hash())
	if err != nil {
		log.LogWarn("收据获取失败，使用估算: %v", err)
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

// SellTokenViaManager2 一级市场卖出全部持仓 (纯 TokenManager2 V2 版本, BNB quote盘子)
func SellTokenViaManager2(target *configentity.MonitorTarget, tokenAddr string, balance *big.Int) (string, error) {
	ethClient := manager.GetEthClient()
	info := client.GetTokenStatus(tokenAddr)
	if info.Status != client.StatusTrading {
		log.LogWarn("Token 状态非 TRADING (%s)，跳过卖出: %s", info.Status, tokenAddr[:10])
		return "", nil
	}

	tokenAddress := common.HexToAddress(tokenAddr)
	walletAddress := common.HexToAddress(config.BSCChain.WalletAddress)

	if balance.Sign() == 0 {
		log.LogWarn("Token %s 余额为0，无需卖出", tokenAddr[:10])
		return "", nil
	}

	log.LogInfo("准备卖出全部 Token: %s | SellAmount: %s", tokenAddr, balance.String())

	// 本地计算 minFunds (滑点保护，预期最低BNB输出)
	minFunds, err := client.LocalCalcMinAmountOut(info, balance, target.SlippageTolerance) // 假设你有这个反向计算方法
	if err != nil {
		log.LogWarn("本地滑点计算失败，使用保守值: %v", err)
		slippagePercent := big.NewInt(int64(target.SlippageTolerance * 100))
		minFunds = new(big.Int).Mul(balance, new(big.Int).Sub(big.NewInt(100), slippagePercent))
		minFunds.Div(minFunds, big.NewInt(100))
	}
	if minFunds.Sign() == 0 {
		minFunds = big.NewInt(1) // 至少1 wei，避免合约revert
	}

	managerABI := utils.GetABI("TokenManager2")
	if managerABI == nil {
		return "", fmt.Errorf("TokenManager2 ABI 未加载")
	}

	tokenABI := utils.GetABI("ERC20") // 标准ERC20 ABI，用于approve/allowance
	if tokenABI == nil {
		return "", fmt.Errorf("ERC20 ABI 未加载")
	}

	managerAddr := common.HexToAddress(config.TokenManager2)

	// 第一步：检查Allowance并自动Approve
	allowanceInput, _ := tokenABI.Pack("allowance", walletAddress, managerAddr)
	callMsg := ethereum.CallMsg{
		From: walletAddress,
		To:   &tokenAddress,
		Data: allowanceInput,
	}
	result, err := ethClient.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return "", fmt.Errorf("查询allowance失败: %v", err)
	}
	currentAllowance := new(big.Int)
	tokenABI.UnpackIntoInterface(&currentAllowance, "allowance", result)

	if currentAllowance.Cmp(balance) < 0 {
		log.LogInfo("Allowance不足 (%s < %s)，执行Approve全额", currentAllowance.String(), balance.String())

		approveInput, err := tokenABI.Pack("approve", managerAddr, balance)
		if err != nil {
			return "", fmt.Errorf("Pack approve 失败: %v", err)
		}

		nonce, _ := ethClient.PendingNonceAt(context.Background(), walletAddress)
		gasPrice, _ := ethClient.SuggestGasPrice(context.Background())

		approveTx := types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			To:       &tokenAddress,
			Value:    big.NewInt(0),
			Gas:      100000,
			GasPrice: gasPrice,
			Data:     approveInput,
		})

		privateKey, err := trade.GetPrivateKey()
		if err != nil {
			return "", err
		}
		signedApproveTx, err := types.SignTx(approveTx, types.NewEIP155Signer(config.BSCChain.ChainID), privateKey)
		if err != nil {
			return "", err
		}

		err = ethClient.SendTransaction(context.Background(), signedApproveTx)
		if err != nil {
			return "", fmt.Errorf("Approve交易发送失败: %v", err)
		}
		log.LogInfo("Approve提交成功 Tx: %s", signedApproveTx.Hash().Hex())

		// 等待approve确认
		_, err = trade.WaitForReceipt(ethClient, signedApproveTx.Hash())
		if err != nil {
			return "", fmt.Errorf("等待Approve确认失败: %v", err)
		}
		log.LogInfo("Approve已确认")
		// approve后nonce已+1，后续卖出会自动使用新nonce
	}

	// 第二步：Pack sellToken (最常用3参数版本)
	input, err := managerABI.Pack("sellToken",
		big.NewInt(0),
		tokenAddress,
		balance,  // amount = 全部余额
		minFunds, // minFunds 滑点保护
		big.NewInt(10),
		walletAddress,
	)
	if err != nil {
		return "", fmt.Errorf("Pack sellToken 失败: %v", err)
	}

	// 构建卖出交易
	nonce, nonceErr := ethClient.PendingNonceAt(context.Background(), walletAddress)
	if nonceErr != nil {
		return "", fmt.Errorf("获取 nonce 失败: %v", nonceErr)
	}
	log.LogInfo("当前钱包 nonce: %d", nonce)

	gasPrice, gasErr := ethClient.SuggestGasPrice(context.Background())
	if gasErr != nil {
		log.LogWarn("获取 gasPrice 失败，使用默认1 gwei: %v", gasErr)
		gasPrice = big.NewInt(1000000000)
	} else {
		gasPrice.Mul(gasPrice, big.NewInt(11)).Div(gasPrice, big.NewInt(10)) // +10%
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &managerAddr,
		Value:    big.NewInt(0), // 非payable
		Gas:      1200000,
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
		database.SaveTxRecord(&po.TransactionRecord{
			Type:      "sell_primary_v2",
			TokenAddr: tokenAddr,
			AmountOut: balance,
			Status:    "failed",
			ErrorMsg:  err.Error(),
			Timestamp: time.Now(),
		})
		return "", fmt.Errorf("发送卖出交易失败: %v", err)
	}

	txHash := signedTx.Hash().Hex()
	log.LogInfo("一级卖出全部提交成功 (V2) Tx: %s Token: %s Amount: %s", txHash, tokenAddr[:10], balance.String())

	// 等待收据 + 更新持仓（减持为0）
	//receipt, err := waitForReceipt(ethClient, signedTx.Hash())
	//if err != nil {
	//	log.LogWarn("收据获取失败，使用估算: %v", err)
	//	// 估算更新持仓（卖出全部，持仓清零）
	//	clearPositionEstimate(tokenAddr, txHash, target) // 你可以自行实现清零逻辑
	//	return txHash, nil
	//}
	//
	//actualFunds := extractFundsOutFromReceipt(receipt, tokenAddr) // 假设有提取BNB数量方法
	//if actualFunds == nil {
	//	actualFunds = minFunds
	//}
	//clearPositionFromReceipt(tokenAddr, txHash, target, actualFunds) // 清零持仓

	return txHash, nil
}
