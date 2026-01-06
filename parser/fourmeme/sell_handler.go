// parser/fourmeme/sell_handler.go
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

type SellHandler struct{}

func NewSellHandler() *SellHandler {
	return &SellHandler{}
}
func (h *SellHandler) ParseLogMulti(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget) {
}
func (h *SellHandler) ParseLog(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget) {
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
		if t0 != config.HashManager1EventSell && t0 != config.HashManager2EventSell {
			continue // 不是 sell
		}

		words := utils.SplitDataToWords(logData.Data)
		if len(words) < 6 {
			continue
		}

		tokenAddr := common.BytesToAddress(words[0].Bytes()).Hex()
		seller := common.BytesToAddress(words[1].Bytes()).Hex()

		var tokenAmount, bnbAmount *big.Int
		var protocol = "managerV2"
		var protocolAddress = config.AddrTokenManager2

		if t0 == config.HashManager1EventSell {
			// Manager1 sell
			tokenAmount = words[2]
			bnbAmount = words[3]
			protocol = "managerV1"
			protocolAddress = config.AddrTokenManager1
		} else {
			// Manager2 sell
			tokenAmount = words[3]
			bnbAmount = words[4]
		}

		_, paymentTokenSymbol, err := getPaymentToken(tokenAddr)
		if err != nil {
			log.LogError("getPaymentToken err: %v", err)
			return
		}

		log.LogInfo("【卖出】Token: %s | Seller: %s | Token: %s | Pay: %s %s | Tx: %s",
			tokenAddr, seller, utils.Div10Pow(tokenAmount, big.NewInt(18)).String(), paymentTokenSymbol, utils.Div10Pow(bnbAmount, big.NewInt(18)).String(), txHash)

		//go func() {
		//	_, err := trade.Sell(target, tokenAddr)
		//	if err != nil {
		//		log.LogError("自动卖出失败: %v", err)
		//	}
		//}()

		// 入库
		saveTradeRecord(blockNum, txHash, txIndex, seller, "", protocol, protocolAddress.Hex(), tokenAddr, paymentTokenSymbol, tokenAmount, bnbAmount, 4)
	}
}
