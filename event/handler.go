// event/handler.go
package event

import (
	"go_fourmeme/manager"
	"math/big"
	"time"

	"go_fourmeme/config"
	"go_fourmeme/database"
	"go_fourmeme/entity/po"
	"go_fourmeme/log"
	"go_fourmeme/trade"
	"go_fourmeme/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	// 标准事件Topic（从交易和标准ABI）
	transferTopic     = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	approvalTopic     = common.HexToHash("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925")
	safeReceivedTopic = common.HexToHash("0x3d0ce9bfc3ed7d6862dbb28b2dea94561fe714a1b4d019aa8af39730d1ad7c3d")

	// Fourmeme自定义事件Topic（从官方交易提取）
	depositConfirmTopic = common.HexToHash("0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19") // 存款/买卖确认
	sigConfirmTopic     = common.HexToHash("0x741ffc4605df23259462547defeab4f6e755bdc5fbb6d0820727d6d3400c7e0d") // 签名/relay确认

	// Pancake事件
	pairCreatedTopic = common.HexToHash("0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c")
	mintTopic        = common.HexToHash("0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9") // Add Liquidity
)

// HandleEvent: 核心事件处理器，使用官方ABI unpack
func HandleEvent(vLog types.Log, target *config.MonitorTarget, client *ethclient.Client) {
	if len(vLog.Topics) == 0 {
		return
	}

	txRecord := &po.TransactionRecord{
		TxHash:    vLog.TxHash.Hex(),
		TokenAddr: vLog.Address.Hex(),
		Timestamp: time.Now(),
		Status:    "info", // 默认info，交易成功/失败会覆盖
	}

	topic0 := vLog.Topics[0]
	isManager := isFourmemeManagerAddr(vLog.Address)

	switch topic0 {
	case transferTopic:
		handleTransfer(vLog, txRecord, target, client)

	case approvalTopic:
		txRecord.Type = "approval"

	case safeReceivedTopic:
		txRecord.Type = "bnb_received" // 流动性相关BNB接收

	case depositConfirmTopic:
		if isManager {
			handleDepositConfirm(vLog, txRecord, target, client)
		}

	case sigConfirmTopic:
		if isManager {
			txRecord.Type = "signature_confirm"
		}

	case pairCreatedTopic:
		handlePairCreated(vLog, txRecord, target, client)

	case mintTopic:
		handleLiquidityAdd(vLog, txRecord, target, client)

	default:
		return // 忽略无关事件
	}

	database.SaveTxRecord(txRecord)
}

// handleTransfer 处理所有 ERC20 Transfer 事件
// 重点识别：
// 1. 用户卖出到 Manager（一级市场卖出）
// 2. Manager 分发 token（一级市场买入确认）
// 3. 聪明钱包买入（可触发跟随买入）
func handleTransfer(vLog types.Log, record *po.TransactionRecord, target *config.MonitorTarget, client *ethclient.Client) {
	record.Type = "transfer"

	var transferEvent struct {
		From  common.Address
		To    common.Address
		Value *big.Int
	}

	// 使用标准 ERC20 ABI 精确 unpack（更可靠）
	if err := utils.ERC20ABI.UnpackIntoInterface(&transferEvent, "Transfer", vLog.Data); err != nil {
		log.LogErrorMsg("Transfer 事件 unpack 失败 (Tx: %s): %v", vLog.TxHash.Hex(), err)
		return
	}

	// indexed 参数从 Topics 取（更准确）
	if len(vLog.Topics) >= 3 {
		transferEvent.From = common.BytesToAddress(vLog.Topics[1].Bytes())
		transferEvent.To = common.BytesToAddress(vLog.Topics[2].Bytes())
	}

	record.AmountOut = transferEvent.Value

	tokenAddr := vLog.Address.Hex() // 当前 Transfer 发生的 token 地址
	fromAddr := transferEvent.From
	toAddr := transferEvent.To
	amount := transferEvent.Value

	// 判断是否涉及 Fourmeme Manager 合约
	isFromManager := isFourmemeManagerAddr(fromAddr)
	isToManager := isFourmemeManagerAddr(toAddr)

	// 1. 用户卖出 token 到 Manager（一级市场卖出）
	if isToManager {
		if isSmartWallet(fromAddr) {
			record.Type = "sell_to_manager_smart"
			log.LogInfo("【聪明钱包一级卖出】%s → Manager | Token: %s | 金额: %s",
				fromAddr.Hex()[:10], tokenAddr[:10], amount.String())
			// 可在这里添加“反向跟随”策略（聪明钱卖我们也卖），目前暂不实现
		} else {
			record.Type = "sell_to_manager_normal"
			log.LogInfo("普通用户一级卖出 → Manager | Token: %s | 金额: %s", tokenAddr[:10], amount.String())
		}
	}

	// 2. Manager 分发 token（一级市场买入确认，或流动性注入后分发）
	if isFromManager {
		record.Type = "buy_from_manager"

		log.LogInfo("【一级市场分发】Manager → %s | Token: %s | 金额: %s",
			toAddr.Hex()[:10], tokenAddr[:10], amount.String())

		// 特别判断：是我们自己的钱包收到 → 说明我们的一级买入成功确认
		if toAddr.String() == config.BSCChain.WalletAddress {
			log.LogInfo("【我们的买入成功确认！】收到 %s %s token 来自 Manager (Tx: %s)",
				amount.String(), tokenAddr[:10], vLog.TxHash.Hex())

			// 可在这里做额外处理（如更新持仓状态、标记交易完成等）
			// 持仓已在 trade.BuyTokenViaManager 成功后添加，此处仅日志确认
		}

		// 如果目标配置了“流动性添加后买入”，且此分发可能是进入二级前的信号，可辅助触发
		// （主要逻辑仍在 Mint/PairCreated 事件）
	}

	// 3. 聪明钱包买入（任意市场买入，可能是一级或二级）
	if isSmartWallet(toAddr) {
		record.Type = "smart_wallet_buy"
		log.LogInfo("【聪明钱包买入】%s 收到 %s token | Token: %s",
			toAddr.Hex()[:10], amount.String(), tokenAddr[:10])

		// 只有在配置了跟随聪明钱包买入时才触发
		if target != nil && target.TriggerOnSmartWalletBuy {
			log.LogInfo("触发跟随买入策略 → 聪明钱包买入 %s，准备买入 Token: %s", amount.String(), tokenAddr[:10])

			// 跟随买入金额可配置为固定值或按比例，这里使用配置中的 BuyAmountBNB
			buyAmountBNB := getBuyAmountInWei(target)

			// 异步执行买入，避免阻塞事件处理
			go func() {
				txHash, err := trade.BuyTokenViaManager(
					manager.GetClient(),
					tokenAddr,
					buyAmountBNB,
					target.SlippageTolerance,
					target,
				)
				if err != nil {
					log.LogErrorMsg("跟随聪明钱包买入失败 (Token: %s): %v", tokenAddr[:10], err)
					return
				}
				log.LogInfo("跟随聪明钱包买入成功提交 Tx: %s", txHash)
			}()
		}
	}

	// 其他普通 transfer 记录为 general_transfer
	if record.Type == "transfer" {
		record.Type = "general_transfer"
	}
}

// handleDepositConfirm: 自定义存款/买卖确认事件（使用TokenManager2 ABI unpack，如果有事件名；否则fallback raw）
func handleDepositConfirm(vLog types.Log, record *po.TransactionRecord, target *config.MonitorTarget, client *ethclient.Client) {
	record.Type = "deposit_confirm"

	// 尝试使用TokenManager2 ABI unpack（假设事件名 "DepositConfirmed" 或类似；如果ABI中有）
	var event struct {
		Token  common.Address
		Sender common.Address
		Amount *big.Int
		// 其他字段...
	}
	if err := utils.GetABI("TokenManager2").UnpackIntoInterface(&event, "DepositConfirmed", vLog.Data); err != nil {
		// fallback raw unpack（从交易示例）
		if len(vLog.Data) >= 96 {
			event.Token = common.BytesToAddress(vLog.Data[12:32]) // padded address
			event.Sender = common.BytesToAddress(vLog.Data[44:64])
			event.Amount = new(big.Int).SetBytes(vLog.Data[64:96])
		} else {
			return
		}
	}

	record.TokenAddr = event.Token.Hex()
	record.AmountOut = event.Amount

	log.LogInfo("Fourmeme存款/买卖确认: Token=%s, Sender=%s, Amount=%s", event.Token.Hex(), event.Sender.Hex(), event.Amount.String())

	// 动态添加新token监听
	addDynamicMonitor(event.Token.Hex())

	// 触发自动买入一级市场
	if target != nil && target.BuyOnLiquidityAdd {
		buyAmount := getBuyAmountInWei(target) // 或基于event.Amount比例
		trade.BuyTokenViaManager(client, event.Token.Hex(), buyAmount, target.SlippageTolerance, target)
	}
}

// 其他handler（pairCreated / liquidityAdd）类似之前，可使用Pancake ABI unpack Sync/Mint

func handlePairCreated(vLog types.Log, record *po.TransactionRecord, target *config.MonitorTarget, client *ethclient.Client) {
	record.Type = "pair_created"
	log.LogInfo("新Pair创建（进入二级）: %s", vLog.TxHash.Hex())
	// unpack PairCreated 获取token0/token1，判断是否目标token
	if target != nil {
		trade.BuyTokenSecondary(client, record.TokenAddr, getBuyAmountInWei(target), target.SlippageTolerance)
	}
}

func handleLiquidityAdd(vLog types.Log, record *po.TransactionRecord, target *config.MonitorTarget, client *ethclient.Client) {
	record.Type = "add_liquidity"
	log.LogInfo("检测到添加流动性")
	if target != nil {
		trade.BuyTokenSecondary(client, record.TokenAddr, getBuyAmountInWei(target), target.SlippageTolerance)
	}
}

// 辅助函数不变
func isFourmemeManagerAddr(addr common.Address) bool {
	managers := []common.Address{
		common.HexToAddress(config.DefaultFourmemeManager),
		common.HexToAddress(config.TokenManagerHelper3),
		common.HexToAddress(config.TokenManager1),
	}
	for _, m := range managers {
		if addr == m {
			return true
		}
	}
	return false
}

func isSmartWallet(addr common.Address) bool {
	for _, w := range config.DefaultSmartWallets.Wallets {
		if common.HexToAddress(w.WalletAddress) == addr {
			return true
		}
	}
	return false
}

// 假设 target.BuyAmountBNB 是 *big.Float，例如 new(big.Float).SetFloat64(0.1)
func getBuyAmountInWei(target *config.MonitorTarget) *big.Int {
	if target == nil || target.BuyAmountBNB == nil {
		return big.NewInt(0)
	}

	// 1 BNB = 10^18 Wei
	weiPerBNB := new(big.Float).SetInt64(1e18)

	// 投入 BNB 数量 * 1e18 → 得到 Wei（float）
	amountInWeiFloat := new(big.Float).Mul(target.BuyAmountBNB, weiPerBNB)

	// 转为 big.Int（截断小数，标准做法）
	amountInWei := new(big.Int)
	amountInWeiFloat.Int(amountInWei) // 注意：是 .Int 方法，不是 .BigInt

	return amountInWei
}

func addDynamicMonitor(tokenAddr string) {
	mu.Lock()
	defer mu.Unlock()
	for _, t := range config.DefaultMonitorTargets {
		if t.TokenAddress == tokenAddr {
			return // 已存在
		}
	}
	newTarget := &config.MonitorTarget{
		TokenName:         "Dynamic_" + tokenAddr[:6],
		TokenAddress:      tokenAddr,
		FourmemeManagers:  config.DefaultMonitorTargets[0].FourmemeManagers, // 继承
		TopicsToMonitor:   config.DefaultTopics,
		BuyOnLiquidityAdd: true,
		// 复制其他策略
	}
	config.DefaultMonitorTargets = append(config.DefaultMonitorTargets, newTarget)
	log.LogInfo("动态添加新token监听: %s", tokenAddr)
	// 注意：需在main中channel通知重启监听，或接受动态query不完美
}
