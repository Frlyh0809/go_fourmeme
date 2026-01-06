// parser/fourmeme/create_handler.go
package fourmeme

import (
	"go_fourmeme/config"
	configentity "go_fourmeme/entity/config"
	"go_fourmeme/log"
	"go_fourmeme/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type CreateHandler struct{}

func NewCreateHandler() *CreateHandler {
	return &CreateHandler{}
}
func (h *CreateHandler) ParseLogMulti(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget) {
}

func (h *CreateHandler) ParseLog(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget) {
	if len(allLogs) == 0 {
		return
	}
	blockNum := receipt.BlockNumber.Uint64()
	txHash := receipt.TxHash.Hex()

	for _, vLog := range allLogs {
		if len(vLog.Topics) < 3 {
			continue
		}
		if vLog.Topics[0] != config.HashOwnershipTransferred {
			continue
		}

		previousOwner := common.BytesToAddress(vLog.Topics[1].Bytes())
		newOwner := common.BytesToAddress(vLog.Topics[2].Bytes())
		tokenAddr := vLog.Address.Hex()

		isCreator, creatorStr := parseCreator(allLogs, vLog.Address)
		realCreator := previousOwner.Hex()
		if isCreator {
			realCreator = creatorStr
		}

		if previousOwner == config.AddrZero {
			log.LogInfo("【新 Token 创建】Token: %s | Creator: %s | blockNum:%d | hash:%s", tokenAddr, realCreator, blockNum, txHash)
		} else if newOwner == config.AddrZero {
			log.LogInfo("【Token 销毁】Token: %s | Owner: %s | blockNum:%d | hash:%s", tokenAddr, previousOwner.Hex(), blockNum, txHash)
		} else if isFourmemeManager(newOwner) {
			log.LogInfo("【Token 移交fourMeme】Token: %s | From: %s | blockNum:%d | hash:%s", tokenAddr, previousOwner.Hex(), blockNum, txHash)
		} else {
			log.LogInfo("【Token Owner 转移】Token: %s | From: %s | To: %s | blockNum:%d | hash:%s", tokenAddr, previousOwner.Hex(), newOwner.Hex(), blockNum, txHash)
		}
		break
	}
}

func parseCreator(allLogs []types.Log, addToken common.Address) (bool, string) {
	if len(allLogs) == 0 {
		return false, ""
	}
	for _, vLog := range allLogs {

		if len(vLog.Topics) == 0 {
			continue

		}
		if vLog.Topics[0] == config.HashTopicManager2CreateEvent1 {
			words := utils.SplitDataToWords(vLog.Data)
			if len(words) < 2 {
				continue
			}
			if addToken == common.BytesToAddress(words[1].Bytes()) {
				creator := common.BytesToAddress(words[0].Bytes())
				return true, creator.Hex()
			}

		}
	}
	return false, "" // 占位
}
