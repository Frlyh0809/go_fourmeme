// parser/fourmeme/buy_handler.go
package fourmeme

import (
	configentity "go_fourmeme/entity/config"
	"math/big"

	"go_fourmeme/config"
	"go_fourmeme/log"
	"go_fourmeme/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type BuyHandler struct{}

func NewBuyHandler() *BuyHandler {
	return &BuyHandler{}
}
func (h *BuyHandler) ParseLogMulti(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget) {
}
func (h *BuyHandler) ParseLog(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget) {
	if len(allLogs) == 0 {
		return
	}
	blockNum := receipt.BlockNumber.Uint64()
	txHash := receipt.TxHash.Hex()
	txIndex := receipt.TransactionIndex

	for _, logData := range allLogs {
		if len(logData.Topics) == 0 {
			continue
		}

		t0 := logData.Topics[0]
		if t0 != config.HashManager1EventBuy && t0 != config.HashManager2EventBuy {
			continue // 不是 buy
		}

		words := utils.SplitDataToWords(logData.Data)
		if len(words) < 6 {
			continue
		}

		tokenAddr := common.BytesToAddress(words[0].Bytes()).Hex()
		buyer := common.BytesToAddress(words[1].Bytes()).Hex()

		var tokenAmount, bnbAmount *big.Int
		var protocol = "managerV2"
		var protocolAddress = config.AddrTokenManager2

		if t0 == config.HashManager1EventBuy {
			// Manager1 buy
			tokenAmount = words[2]
			bnbAmount = words[3]
			protocol = "managerV1"
			protocolAddress = config.AddrTokenManager1
		} else {
			// Manager2 buy
			tokenAmount = words[3]
			bnbAmount = words[4]
		}

		_, paymentTokenSymbol, err := getPaymentToken(tokenAddr)
		if err != nil {
			log.LogError("getPaymentToken err: %v", err)
			return
		}

		log.LogInfo("【买入】Token: %s | Buyer: %s | Token: %s | Pay: %s %s | Tx: %s",
			tokenAddr, buyer, utils.Div10Pow(tokenAmount, big.NewInt(18)).String(), paymentTokenSymbol, utils.Div10Pow(bnbAmount, big.NewInt(18)).String(), txHash)

		//if target != nil {
		//	go func() {
		//		_, err := trade.Buy(target, tokenAddr)
		//		if err != nil {
		//			log.LogError("自动买入失败: %v", err)
		//		}
		//	}()
		//}

		// 入库
		saveTradeRecord(blockNum, txHash, txIndex, "", buyer, protocol, protocolAddress.Hex(), tokenAddr, paymentTokenSymbol, tokenAmount, bnbAmount, 3)
	}
}
