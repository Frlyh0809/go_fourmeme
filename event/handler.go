// event/handler.go
package event

import (
	"fmt"
	"go_fourmeme/client"
	"go_fourmeme/database"
	"go_fourmeme/trade"
	"math/big"
	"strconv"
	"time"

	"go_fourmeme/config"
	configentity "go_fourmeme/entity/config"
	"go_fourmeme/entity/po"
	"go_fourmeme/log"
	"go_fourmeme/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// HandleEventV2 从整个区块日志中解析所有 Fourmeme 事件（创建 + 买卖 + 入库）
func HandleEventV2(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget) {
	if len(allLogs) == 0 {
		return
	}

	blockNum := receipt.BlockNumber.Uint64()
	txHash := receipt.TxHash.Hex()
	txIndex := receipt.TransactionIndex

	// 1. Token 创建/销毁/转移识别 (OwnershipTransferred)
	for _, vLog := range allLogs {
		if len(vLog.Topics) == 0 {
			continue
		}
		if vLog.Topics[0] == config.HashOwnershipTransferred && len(vLog.Topics) >= 3 {
			previousOwner := common.BytesToAddress(vLog.Topics[1].Bytes())
			newOwner := common.BytesToAddress(vLog.Topics[2].Bytes())
			tokenAddr := vLog.Address.Hex()

			// 解析真实 creator
			isCreator, creator := parseCreator(allLogs, vLog.Address)
			realCreator := previousOwner.Hex() // 默认 previousOwner
			if isCreator {
				realCreator = creator
			}

			var recordType int
			var send, receive string

			if previousOwner == config.AddrZero {
				// 创建
				recordType = 0
				send = previousOwner.Hex()
				receive = newOwner.Hex()
				log.LogInfo("【新 Token 创建】Token: %s | Creator: %s | blockNum:%d | hash:%s ", tokenAddr, realCreator, blockNum, txHash)
			} else if newOwner == config.AddrZero {
				// 销毁
				recordType = 8
				send = previousOwner.Hex()
				receive = newOwner.Hex()
				log.LogInfo("【Token 销毁】Token: %s | Creator: %s | blockNum:%d | hash:%s ", tokenAddr, realCreator, blockNum, txHash)
			} else if isFourmemeManager(newOwner) {
				// 移交 fourMeme (owner 转移)
				recordType = 7
				send = previousOwner.Hex()
				receive = newOwner.Hex()
				log.LogInfo("【Token 移交fourMeme】Token: %s | Creator: %s | blockNum:%d | hash:%s ", tokenAddr, realCreator, blockNum, txHash)
			} else {
				// 其他转移
				recordType = 7
				send = previousOwner.Hex()
				receive = newOwner.Hex()
				log.LogInfo("【Token Owner 转移】Token: %s | From: %s | To: %s | blockNum:%d | hash:%s ", tokenAddr, previousOwner.Hex(), newOwner.Hex(), blockNum, txHash)
			}

			// 入库 transaction_create
			createRecord := &po.TransactionCreate{
				TxUniqueSeq:     utils.CalcTxUniqueSeq(blockNum, vLog.TxIndex, vLog.Index),
				BlockNumber:     strconv.FormatUint(blockNum, 10),
				TxIndex:         fmt.Sprintf("%d", txIndex),
				TxHash:          txHash,
				Send:            send,
				Receive:         receive,
				Type:            recordType,
				Protocol:        "managerV2", // 根据实际判断 v1/v2/helper3
				ProtocolAddress: vLog.Address.Hex(),
				TokenAddress:    tokenAddr,
				PaymentToken:    "", // 创建/销毁无支付
				TokenAmount:     "",
				PaymentAmount:   "",
				Price:           "",
				Volume:          "",
				CreatedAt:       time.Now(),
			}
			if err := database.SaveTransactionCreate(createRecord); err != nil {
				log.LogError("保存创建/转移记录失败: %v", err)
			}
			break
		}
	}

	// 2. 买卖识别：匹配自定义事件
	var tradeInfo struct {
		TokenAddr    common.Address
		Buyer        common.Address
		TokenAmount  *big.Int
		BNBAmount    *big.Int
		BNBFeeAmount *big.Int
		IsUSD1       bool
		TxHash       common.Hash
	}

	foundCustom := false
	var protocol = "managerV2"
	var protocolAddress = config.DefaultFourmemeManagerAddr
	for _, logData := range allLogs {
		if len(logData.Topics) == 0 {
			continue
		}

		t0 := logData.Topics[0]
		if t0 == config.HashManager1Event1 || t0 == config.HashManager1Event2 ||
			t0 == config.HashManager2Event1 || t0 == config.HashManager2Event2 {

			words := utils.SplitDataToWords(logData.Data)

			tradeInfo.TokenAddr = common.BytesToAddress(words[0].Bytes())
			tradeInfo.Buyer = common.BytesToAddress(words[1].Bytes())
			if t0 == config.HashManager1Event1 || t0 == config.HashManager1Event2 {
				if len(words) < 5 {
					continue
				}
				tradeInfo.TokenAmount = words[2]
				tradeInfo.BNBAmount = words[3]
				tradeInfo.BNBFeeAmount = words[4]

				protocol = "managerV1"
				protocolAddress = config.AddrTokenManager1
			} else { // Manager2 / Helper3
				if len(words) < 6 {
					continue
				}
				tradeInfo.TokenAmount = words[3]
				tradeInfo.BNBAmount = words[4]
				tradeInfo.BNBFeeAmount = words[5]
			}

			tradeInfo.TxHash = logData.TxHash

			// 检测 USD1 支付
			for _, l := range allLogs {
				if l.Address == config.AddrUSD1 && l.Topics[0] == config.HashTransfer {
					if len(l.Topics) == 3 {
						if common.BytesToAddress(words[1].Bytes()) == config.AddrTokenManagerHelper3 || common.BytesToAddress(words[2].Bytes()) == config.AddrTokenManagerHelper3 {
							tradeInfo.IsUSD1 = true
							protocol = "managerV3"
							protocolAddress = config.AddrTokenManagerHelper3
							break
						}
					}
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
	for _, vLog := range allLogs {
		if vLog.Topics[0] == config.HashTransfer && vLog.Address == tradeInfo.TokenAddr {
			from := common.BytesToAddress(vLog.Topics[1].Bytes())
			to := common.BytesToAddress(vLog.Topics[2].Bytes())
			value := new(big.Int).SetBytes(vLog.Data)

			//log.LogInfo("【Token 交易】Token: %s | trader: %s | TokenAmount:%d  | BNBAmount:%d | value:%d | hash:%s ",
			//	tradeInfo.TokenAddr.Hex(), tradeInfo.Buyer.Hex(), tradeInfo.TokenAmount, tradeInfo.BNBAmount, value, txHash)

			if value.Cmp(tradeInfo.TokenAmount) != 0 {
				continue
			}

			payment := "BNB"
			paymentToken := config.ZeroAddress
			if tradeInfo.IsUSD1 {
				payment = "USD1"
				paymentToken = config.USD1Address
			}

			var recordType int
			var send, receive string

			if isFourmemeManager(to) {
				// 卖出
				recordType = 4
				send = from.Hex()
				receive = to.Hex()
				log.LogInfo("【卖出】Token: %s | Seller: %s | Token: %s | Pay: %s %s | Tx: %s",
					tradeInfo.TokenAddr.Hex(), tradeInfo.Buyer.Hex(), tradeInfo.TokenAmount.String(),
					payment, tradeInfo.BNBAmount.String(), tradeInfo.TxHash.Hex())

			} else if isFourmemeManager(from) {
				// 买入
				recordType = 3
				send = from.Hex()
				receive = to.Hex()
				log.LogInfo("【买入】Token: %s | Buyer: %s | Token: %s | Pay: %s %s | Tx: %s",
					tradeInfo.TokenAddr.Hex(), tradeInfo.Buyer.Hex(), tradeInfo.TokenAmount.String(),
					payment, tradeInfo.BNBAmount.String(), tradeInfo.TxHash.Hex())

				if target != nil {
					//TODO 交易
					//go trade.BuyTokenViaManager(target, tradeInfo.TokenAddr.Hex())
				}
			} else {
				// 普通 Transfer
				recordType = 5
				send = from.Hex()
				receive = to.Hex()
				log.LogInfo("【Transfer】Token: %s | From: %s | To: %s | Amount: %s | Tx: %s",
					tradeInfo.TokenAddr.Hex(), from.Hex(), to.Hex(),
					tradeInfo.TokenAmount.String(), tradeInfo.TxHash.Hex())
			}
			bnbPrice := client.GetBNBPriceUSDT()
			tokenAmountDis := utils.Div10Pow(tradeInfo.TokenAmount, big.NewInt(18))
			bnbAmountDis := utils.Div10Pow(tradeInfo.BNBAmount, big.NewInt(18))
			price := utils.DivFloat(bnbAmountDis, tokenAmountDis)
			volume := bnbAmountDis
			if payment == "USD1" {
				//单位使用bnb
				price = new(big.Float).Mul(price, big.NewFloat(bnbPrice))
				volume = new(big.Float).Mul(volume, big.NewFloat(bnbPrice))
			}

			// 入库 transaction
			txRecord := &po.Transaction{
				TxUniqueSeq:     utils.CalcTxUniqueSeq(blockNum, vLog.TxIndex, vLog.Index),
				BlockNumber:     strconv.FormatUint(blockNum, 10),
				TxIndex:         fmt.Sprintf("%d", vLog.TxIndex),
				TxHash:          txHash,
				Send:            send,
				Receive:         receive,
				Type:            recordType,
				Protocol:        protocol, // 根据实际判断
				ProtocolAddress: protocolAddress.Hex(),
				TokenAddress:    tradeInfo.TokenAddr.Hex(),
				PaymentToken:    paymentToken,
				TokenAmount:     utils.BigFloatToString(tokenAmountDis), // string
				PaymentAmount:   utils.BigFloatToString(bnbAmountDis),   // string
				Price:           utils.BigFloatToString(price),
				Volume:          utils.BigFloatToString(volume),
				CreatedAt:       time.Now(),
			}
			if err := database.SaveTransaction(txRecord); err != nil {
				log.LogError("保存 transaction 记录失败: %v", err)
			}

			break
		}
	}
}

// HandleEventV2 完整买卖解析（结合自定义事件 + Transfer）
func HandleEventV1(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget) {
	if len(allLogs) == 0 {
		return
	}
	blockNum := receipt.BlockNumber
	txHash := receipt.TxHash.Hex()

	// 1. Token 创建识别 (OwnershipTransferred + previousOwner == zero)
	for _, vLog := range allLogs {
		if len(vLog.Topics) == 0 {
			continue
		}
		if vLog.Topics[0] == config.HashOwnershipTransferred && len(vLog.Topics) >= 3 {
			previousOwner := common.BytesToAddress(vLog.Topics[1].Bytes())
			if previousOwner == config.AddrZero {

				newOwner := common.BytesToAddress(vLog.Topics[2].Bytes())
				//解析真实的creator
				isCreator, creator := parseCreator(allLogs, vLog.Address)

				var realCreator common.Address
				if isCreator {
					realCreator = common.HexToAddress(creator)
				} else {
					realCreator = newOwner
					//不是真实的fourmeme 创建
					//continue
				}

				log.LogInfo("【新 Token 创建】Token: %s | Creator: %s | blockNum:%d | hash:%s ", vLog.Address.Hex(), realCreator, blockNum, txHash)
				// 可动态添加监听
				break
			}
			newOwner := common.BytesToAddress(vLog.Topics[2].Bytes())
			if newOwner == config.AddrZero {
				log.LogInfo("【Token 销毁】Token: %s | Creator: %s | blockNum:%d | hash:%s ", vLog.Address.Hex(), previousOwner.Hex(), blockNum, txHash)
				break
			}
			if isFourmemeManager(newOwner) {
				log.LogInfo("【Token 移交fourMeme】Token: %s | Creator: %s | blockNum:%d | hash:%s ", vLog.Address.Hex(), previousOwner.Hex(), blockNum, txHash)
				break
			}
		}
	}

	// 2. 买卖识别：匹配自定义事件
	var tradeInfo struct {
		TokenAddr    common.Address
		Buyer        common.Address
		TokenAmount  *big.Int
		BNBAmount    *big.Int
		BNBFeeAmount *big.Int
		IsUSD1       bool
		TxHash       common.Hash
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
			if t0 == config.HashManager1Event1 || t0 == config.HashManager1Event2 {
				tradeInfo.TokenAmount = words[2]
				tradeInfo.BNBAmount = words[3]
				tradeInfo.BNBFeeAmount = words[4]
			}
			if t0 == config.HashManager2Event1 || t0 == config.HashManager2Event2 {
				tradeInfo.TokenAmount = words[3]
				tradeInfo.BNBAmount = words[4]
				tradeInfo.BNBFeeAmount = words[5]

			}

			tradeInfo.TxHash = logData.TxHash

			// 检测 USD1 支付
			for _, l := range allLogs {
				if len(l.Topics) == 0 {
					continue
				}
				if l.Address == config.AddrUSD1 && l.Topics[0] == config.HashTransfer {
					if len(l.Topics) == 3 {
						if common.BytesToAddress(words[1].Bytes()) == config.AddrTokenManagerHelper3 || common.BytesToAddress(words[2].Bytes()) == config.AddrTokenManagerHelper3 {
							tradeInfo.IsUSD1 = true
							break
						}
					}

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
		if len(logData.Topics) == 0 {
			continue
		}
		if logData.Topics[0] == config.HashTransfer && logData.Address == tradeInfo.TokenAddr {
			from := common.BytesToAddress(logData.Topics[1].Bytes())
			to := common.BytesToAddress(logData.Topics[2].Bytes())
			value := new(big.Int).SetBytes(logData.Data)

			//log.LogInfo("【Token 交易】Token: %s | trader: %s | TokenAmount:%d  | BNBAmount:%d | value:%d | hash:%s ",
			//	tradeInfo.TokenAddr.Hex(), tradeInfo.Buyer.Hex(), tradeInfo.TokenAmount, tradeInfo.BNBAmount, value, txHash)

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
					tradeInfo.TokenAddr.Hex(), tradeInfo.Buyer.Hex(), tradeInfo.TokenAmount.String(),
					payment, tradeInfo.BNBAmount.String(), tradeInfo.TxHash.Hex())
				// 可触发卖出逻辑
			} else if isFourmemeManager(from) {
				// Manager → User = 买入
				log.LogInfo("【买入】Token: %s | Buyer: %s | Token: %s | Pay: %s %s | Tx: %s",
					tradeInfo.TokenAddr.Hex(), tradeInfo.Buyer.Hex(), tradeInfo.TokenAmount.String(),
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
