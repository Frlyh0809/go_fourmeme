// event/handler.go
package event

import (
	"go_fourmeme/database"
	"math/big"
	"time"

	"go_fourmeme/config"
	configentity "go_fourmeme/entity/config"
	"go_fourmeme/entity/po"
	"go_fourmeme/log"
	"go_fourmeme/trade"
	"go_fourmeme/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// HandleEventV2 完整买卖解析（结合自定义事件 + Transfer）
func HandleEventV2(allLogs []types.Log, target *configentity.MonitorTarget) {
	if len(allLogs) == 0 {
		return
	}

	// 1. Token 创建识别 (OwnershipTransferred + previousOwner == zero)
	for _, vLog := range allLogs {
		if vLog.Topics[0] == config.HashOwnershipTransferred && len(vLog.Topics) >= 3 {
			previousOwner := common.BytesToAddress(vLog.Topics[1].Bytes())
			if previousOwner == config.AddrZero {
				newOwner := common.BytesToAddress(vLog.Topics[2].Bytes())
				log.LogInfo("【新 Token 创建】Token: %s | Creator: %s", vLog.Address.Hex(), newOwner.Hex())
				// 可动态添加监听
				return
			}
			newOwner := common.BytesToAddress(vLog.Topics[2].Bytes())
			if newOwner == config.AddrZero {
				log.LogInfo("【Token 销毁】Token: %s | Creator: %s", vLog.Address.Hex(), previousOwner.Hex())
				return
			}
			if isFourmemeManager(newOwner) {
				log.LogInfo("【Token 移交fourMeme】Token: %s | Creator: %s", vLog.Address.Hex(), previousOwner.Hex())
				return
			}
		}
	}

	// 2. 买卖识别：匹配自定义事件
	var tradeInfo struct {
		TokenAddr   common.Address
		Buyer       common.Address
		TokenAmount *big.Int
		BNBAmount   *big.Int
		IsUSD1      bool
		TxHash      common.Hash
	}

	foundCustom := false
	for _, logData := range allLogs {
		if len(logData.Topics) == 0 {
			continue
		}

		t0 := logData.Topics[0]
		if t0 == config.HashManager1Event1 || t0 == config.HashManager1Event2 ||
			t0 == config.HashManager2Event1 || t0 == config.HashManager2Event2 {

			words := utils.SplitDataToWords(logData.Data)
			if len(words) < 5 { // 至少需要 token, buyer, tokenAmt, bnbAmt, ...
				continue
			}

			// 索引含义（根据你的解析文档调整）
			tradeInfo.TokenAddr = common.BytesToAddress(words[0].Bytes()) // 第1个 word (padded address)
			tradeInfo.Buyer = common.BytesToAddress(words[1].Bytes())     // 第2个
			tradeInfo.TokenAmount = words[2]
			tradeInfo.BNBAmount = words[3]
			tradeInfo.TxHash = logData.TxHash

			// 检测 USD1 支付
			for _, l := range allLogs {
				if l.Address == config.AddrUSD1 && l.Topics[0] == config.HashTransfer {
					tradeInfo.IsUSD1 = true
					break
				}
			}

			foundCustom = true
			break
		}
	}

	if !foundCustom {
		return // 无自定义事件，不处理
	}

	// 3. 用 Transfer 事件确认方向
	for _, logData := range allLogs {
		if logData.Topics[0] == config.HashTransfer && logData.Address == tradeInfo.TokenAddr {
			from := common.BytesToAddress(logData.Topics[1].Bytes())
			to := common.BytesToAddress(logData.Topics[2].Bytes())
			value := new(big.Int).SetBytes(logData.Data)

			if value.Cmp(tradeInfo.TokenAmount) != 0 {
				continue
			}

			payment := "BNB"
			if tradeInfo.IsUSD1 {
				payment = "USD1"
			}

			if isFourmemeManager(to) {
				// Token → Manager = 卖出
				log.LogInfo("【卖出】Token: %s | Seller: %s | Token: %s | Pay: %s %s | Tx: %s",
					tradeInfo.TokenAddr.Hex()[:10], from.Hex()[:10], tradeInfo.TokenAmount.String(),
					payment, tradeInfo.BNBAmount.String(), tradeInfo.TxHash.Hex())
				// 可触发卖出逻辑
			} else if isFourmemeManager(from) {
				// Manager → User = 买入
				log.LogInfo("【买入】Token: %s | Buyer: %s | Token: %s | Pay: %s %s | Tx: %s",
					tradeInfo.TokenAddr.Hex()[:10], to.Hex()[:10], tradeInfo.TokenAmount.String(),
					payment, tradeInfo.BNBAmount.String(), tradeInfo.TxHash.Hex())

				if target != nil {
					//go trade.BuyTokenViaManager(target, tradeInfo.TokenAddr.Hex())
				}
			}
			break
		}
	}
}

// HandleEvent 核心事件处理器
func HandleEvent(vLog types.Log, target *configentity.MonitorTarget) {
	if len(vLog.Topics) == 0 {
		return
	}

	record := &po.TransactionRecord{
		TxHash:    vLog.TxHash.Hex(),
		TokenAddr: vLog.Address.Hex(),
		Timestamp: time.Now(),
		Status:    "info",
	}

	topic0 := vLog.Topics[0].Hex()

	switch topic0 {
	case config.TopicTransfer:
		handleTransfer(vLog, record, target)
	case config.PairCreatedTopic:
		handlePairCreated(vLog, record, target)
	case config.MintTopic:
		handleLiquidityAdd(vLog, record, target)
	case config.DepositConfirmTopic:
		handleDepositConfirm(vLog, record, target)
	default:
		// 其他 Fourmeme 或 Pancake 事件可扩展
		return
	}

	err := database.SaveTxRecord(record)
	if err != nil {
		return
	}
}

// handleTransfer 处理 Transfer 事件
func handleTransfer(vLog types.Log, record *po.TransactionRecord, target *configentity.MonitorTarget) {
	record.Type = "transfer"

	if len(vLog.Topics) < 3 {
		return
	}

	var event struct {
		From  common.Address
		To    common.Address
		Value *big.Int
	}
	if err := utils.ERC20ABI.UnpackIntoInterface(&event, "Transfer", vLog.Data); err != nil {
		log.LogError("Transfer unpack 失败: %v", err)
		return
	}
	if len(vLog.Topics) >= 3 {
		event.From = common.BytesToAddress(vLog.Topics[1].Bytes())
		event.To = common.BytesToAddress(vLog.Topics[2].Bytes())
	}

	record.AmountOut = event.Value

	isFromManager := isFourmemeManager(event.From)
	isToManager := isFourmemeManager(event.To)

	if isToManager {
		record.Type = "sell_to_manager"
		if isSmartWallet(event.From) {
			record.Type = "sell_to_manager_smart"
		}
	}

	if isFromManager {
		record.Type = "buy_from_manager"
	}

	if isSmartWallet(event.To) && target.TriggerOnSmartWalletBuy {
		record.Type = "smart_wallet_buy"
		go trade.BuyTokenViaManager(target, vLog.Address.Hex())
	}
}

// handleLiquidityAdd 处理添加流动性（进入二级）
func handleLiquidityAdd(vLog types.Log, record *po.TransactionRecord, target *configentity.MonitorTarget) {
	record.Type = "add_liquidity"
	log.LogInfo("检测到添加流动性，Token: %s", vLog.Address.Hex())

	if target.BuyOnLiquidityAdd {
		go trade.BuyTokenSecondary(target, vLog.Address.Hex())
	}
}

// handlePairCreated 处理新 Pair 创建
func handlePairCreated(vLog types.Log, record *po.TransactionRecord, target *configentity.MonitorTarget) {
	record.Type = "pair_created"
	log.LogInfo("新 Pair 创建: %s", vLog.TxHash.Hex())
	// 可解析 token0/token1，动态添加 token 监听
}

// handleDepositConfirm 处理 Fourmeme 存款确认
func handleDepositConfirm(vLog types.Log, record *po.TransactionRecord, target *configentity.MonitorTarget) {
	record.Type = "deposit_confirm"
	// raw unpack 或 ABI unpack
	// 触发一级买入
	go trade.BuyTokenViaManager(target, vLog.Address.Hex())
}

// 辅助函数
func isFourmemeManager(addr common.Address) bool {
	return addr == config.DefaultFourmemeManagerAddr || addr == common.HexToAddress(config.TokenManagerHelper3) || addr == common.HexToAddress(config.TokenManager1)
}

func isSmartWallet(addr common.Address) bool {
	// 从 config.DefaultSmartWallets 实现
	return false // 占位
}
