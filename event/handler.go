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

var (
	transferTopic       = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	pairCreatedTopic    = "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c"
	mintTopic           = "0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9"
	depositConfirmTopic = "0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19"
)

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
	case transferTopic:
		handleTransfer(vLog, record, target)
	case pairCreatedTopic:
		handlePairCreated(vLog, record, target)
	case mintTopic:
		handleLiquidityAdd(vLog, record, target)
	case depositConfirmTopic:
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

	isFromManager := isFourmemeManagerAddr(event.From)
	isToManager := isFourmemeManagerAddr(event.To)

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
func isFourmemeManagerAddr(addr common.Address) bool {
	return addr == config.DefaultFourmemeManagerAddr || addr == common.HexToAddress(config.TokenManagerHelper3) || addr == common.HexToAddress(config.TokenManager1)
}

func isSmartWallet(addr common.Address) bool {
	// 从 config.DefaultSmartWallets 实现
	return false // 占位
}
