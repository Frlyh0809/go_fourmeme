package trade

import (
	"go_fourmeme/client"
	"go_fourmeme/config"
	"go_fourmeme/database"
	"go_fourmeme/entity"
	"go_fourmeme/entity/po"
	"time"

	configentity "go_fourmeme/entity/config"
	//"go_fourmeme/entity/po"
	"go_fourmeme/log"
	"go_fourmeme/manager"
	"go_fourmeme/trade/internal_market/v2" // 只导入子包执行交易
	v3 "go_fourmeme/trade/internal_market/v3"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// PrimaryBuy 一级市场买入 (计算参数 + 路由 + 后处理)
func PrimaryBuy(target *configentity.MonitorTarget, tokenAddr string, info client.TokenInfo) (string, error) {
	// 计算参数
	amountInWei := ToWei(target.BuyAmountBNB)
	log.LogInfo("-->[准备买入]token:%s buyAmountBNB:%f amountInWei:%s", tokenAddr, target.BuyAmountBNB, amountInWei)
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
		txHash, err = v3.BuyTokenViaManagerV3(tokenAddr, amountInWei, minAmountOut, privateKey)

	} else {
		if info.TokenManager == config.AddrTokenManager1 {
			//v1.
		}
		//v2.
		txHash, err = v2.BuyTokenViaManagerV2(tokenAddr, amountInWei, minAmountOut, privateKey) // 子包函数
	}

	if err != nil {
		log.LogError("一级买入失败: %v1", err)
		database.SaveTxRecord(&po.TransactionRecord{
			Type:      "buy_primary",
			TokenAddr: tokenAddr,
			AmountIn:  amountInWei.String(),
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
	manager.AddPosition(&entity.Position{
		TokenAddr:      tokenAddr,           //string     `json:"token_addr"`
		BuyTxHash:      txHash,              //string     `json:"buy_tx_hash"`
		BuyAmountBNB:   target.BuyAmountBNB, //*big.Float `json:"buy_amount_bnb"`   // 投入 BNB 数量
		BuyTokenAmount: actualOut,           //*big.Int   `json:"buy_token_amount"` // 买入 token 数量
		//BuyPriceAvg:      ,//*big.Float `json:"buy_price_avg"`    // 平均买入价格 (BNB/token)
		//BuyTime:          ,//time.Time  `json:"buy_time"`
		//TargetProfitMult: ,//float64    `json:"target_profit_mult"` // 止盈倍数
		//TargetLossMult:   ,//float64    `json:"target_loss_mult"`   // 止损倍数
		//Sold:             ,//bool       `json:"sold"`
	}) // 示例
	database.SaveTxRecord(&po.TransactionRecord{
		Type:      "buy_primary",
		TokenAddr: tokenAddr,
		AmountIn:  amountInWei.String(),
		Status:    "success",
		TxHash:    txHash,
		Timestamp: time.Now(),
	})

	return txHash, nil
}

// primarySell 一级市场卖出 (计算参数 + 路由 + 后处理)
func primarySell(target *configentity.MonitorTarget, tokenAddr string, info client.TokenInfo) (string, error) {
	// 计算参数: 查询余额作为sellAmount
	//tokenAddress := common.HexToAddress(tokenAddr)
	//walletAddress := common.HexToAddress(config.BSCChain.WalletAddress)

	tokenPositions := manager.GetAllPositions()[tokenAddr]

	if tokenPositions == nil {
		return "", nil
	}

	holdTokenAmounts := tokenPositions.BuyTokenAmount
	minFunds, err := client.LocalCalcMinAmountOut(info, holdTokenAmounts, target.SlippageTolerance)
	log.LogInfo("-->[准备卖出]token:%s holdTokenAmounts:%f amountInWei:%s", tokenAddr, holdTokenAmounts, minFunds)

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
		//v3. TODO 没有走通
		txHash, err = v3.SellTokenViaManagerV3(tokenAddr, holdTokenAmounts, minFunds, privateKey)
	} else {
		if info.TokenManager == config.AddrTokenManager1 {
			//v1.
		}
		txHash, err = v2.SellTokenViaManagerV2(tokenAddr, holdTokenAmounts, minFunds, privateKey) // 子包函数

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
