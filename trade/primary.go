package trade

import (
	"go_fourmeme/client"
	"go_fourmeme/config"
	"go_fourmeme/database"
	configentity "go_fourmeme/entity/config"
	"go_fourmeme/entity/po"
	"go_fourmeme/log"
	"go_fourmeme/manager"
	"go_fourmeme/trade/internal_market/v2" // 只导入子包执行交易
	v3 "go_fourmeme/trade/internal_market/v3"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// PrimaryBuy 一级市场买入 (计算参数 + 路由 + 后处理)
func PrimaryBuy(target *configentity.MonitorTarget, tokenAddr string) (string, error) {
	info, err := client.GetTokenStatus(tokenAddr)
	if err != nil {
		return "", err
	}
	// 计算参数
	amountInWei := ToWei(target.BuyAmountBNB)
	minAmountOut, err := client.LocalCalcMinAmountOut(info, amountInWei, target.SlippageTolerance)
	if err != nil {
		log.LogError("LocalCalcMinAmountOut 本地滑点计算最低获得的token失败 手动设置：0;  err:%v1", err)
		minAmountOut = big.NewInt(1)
	}
	if minAmountOut.Sign() == 0 {
		minAmountOut = big.NewInt(1)
	}
	privateKey, err := GetPrivateKey()
	if err != nil {
		return "", err
	}
	// 路由到对应版本
	var txHash string

	if info.Quote == config.AddrUSDT || info.Quote == config.AddrUSD1 {
		txHash, err = v2.BuyTokenViaManagerV2(tokenAddr, amountInWei, minAmountOut, privateKey) // 子包函数
	} else {
		if info.TokenManager == config.AddrTokenManager1 {
			//v1.
		}
		//v3.
		txHash, err = v3.BuyTokenViaManagerV3(tokenAddr, amountInWei, minAmountOut, privateKey)
	}

	if err != nil {
		log.LogError("一级买入失败: %v1", err)
		database.SaveTxRecord(&po.TransactionRecord{
			Type:      "buy_primary",
			TokenAddr: tokenAddr,
			AmountIn:  amountInWei,
			Status:    "failed",
			ErrorMsg:  err.Error(),
			Timestamp: time.Now(),
		})
		return "", err
	}

	ethClient := manager.GetEthClient()
	receipt, err := WaitForReceipt(ethClient, common.HexToHash(txHash))
	if err != nil {
		log.LogWarn("收据获取失败，使用估算: %v1", err)
		//ClearPositionFromEstimate(tokenAddr, txHash, target, holdTokenAmounts)
		return txHash, nil
	}
	actualOut := ExtractTokenOutFromReceipt(receipt, tokenAddr)
	if actualOut == nil {
		actualOut = minAmountOut
	}
	// 后处理: 更新manager、入库
	//manager.UpdatePosition(tokenAddr, txHash) // 示例
	//database.SaveTxRecord(&po.TransactionRecord{
	//	Type:      "buy_primary",
	//	TokenAddr: tokenAddr,
	//	AmountIn:  amountInWei,
	//	Status:    "success",
	//	TxHash:    txHash,
	//	Timestamp: time.Now(),
	//})

	return txHash, nil
}

// primarySell 一级市场卖出 (计算参数 + 路由 + 后处理)
func primarySell(target *configentity.MonitorTarget, tokenAddr string) (string, error) {
	info, err := client.GetTokenStatus(tokenAddr)
	if err != nil {
		return "", err
	}
	// 计算参数: 查询余额作为sellAmount
	//tokenAddress := common.HexToAddress(tokenAddr)
	//walletAddress := common.HexToAddress(config.BSCChain.WalletAddress)

	holdTokenAmounts := manager.GetAllPositions()[tokenAddr].BuyTokenAmount
	minFunds, err := client.LocalCalcMinAmountOut(info, holdTokenAmounts, target.SlippageTolerance)
	if err != nil {
		log.LogError("LocalCalcMinAmountOut 本地滑点计算最低获得的bnb失败 手动设置：0;  err:%v1", err)
		minFunds = big.NewInt(1)
	}
	if minFunds.Sign() == 0 {
		minFunds = big.NewInt(1)
	}
	privateKey, err := GetPrivateKey()
	if err != nil {
		return "", err
	}
	// 路由到对应版本
	var txHash string
	if info.Quote == config.AddrUSDT || info.Quote == config.AddrUSD1 {
		_, err := client.Approve(tokenAddr, config.TokenManager2, holdTokenAmounts, privateKey)
		if err != nil {
			log.LogError("primarySell 授权失败 %v1", err)
			return "", err
		}
		txHash, err = v2.SellTokenViaManagerV2(tokenAddr, holdTokenAmounts, minFunds, privateKey) // 子包函数
	} else {
		if info.TokenManager == config.AddrTokenManager1 {
			//v1.
		}
		//v3.
	}

	//ethClient := manager.GetEthClient()
	//receipt, err := WaitForReceipt(ethClient, common.HexToHash(txHash))
	//if err != nil {
	//	log.LogWarn("收据获取失败，使用估算: %v1", err)
	//	//ClearPositionFromEstimate(tokenAddr, txHash, target, holdTokenAmounts)
	//	return txHash, nil
	//}
	//actualOut := ExtractTokenOutFromReceipt(receipt, tokenAddr)
	//if actualOut == nil {
	//	actualOut = minAmountOut
	//}

	// 后处理: 更新manager、入库，清仓
	//manager.ClearPosition(tokenAddr, txHash) // 示例
	//database.SaveTxRecord(&po.TransactionRecord{
	//	Type:      "sell_primary",
	//	TokenAddr: tokenAddr,
	//	AmountOut: balance,
	//	Status:    "success",
	//	TxHash:    txHash,
	//	Timestamp: time.Now(),
	//})

	return txHash, nil
}
