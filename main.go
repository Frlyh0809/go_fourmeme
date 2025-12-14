// main.go
package main

import (
	"context"
	"go_fourmeme/entity"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"go_fourmeme/config"
	"go_fourmeme/database"
	"go_fourmeme/event"
	"go_fourmeme/log"
	"go_fourmeme/manager"
	"go_fourmeme/trade"
	"go_fourmeme/utils"
)

func main() {
	// 1. 加载配置（env + yaml/json）
	config.LoadFromEnvAndFile("config.yaml")

	// 2. 初始化日志系统
	log.InitLogger()

	// 3. 从根目录加载所有 ABI 文件
	if err := utils.LoadABIs(); err != nil {
		log.LogFatal("ABI 文件加载失败: %v", err)
	}

	// 4. 初始化数据库
	database.InitDB(config.BSCChain)

	// 5. 连接 BSC WebSocket（用于实时事件订阅）
	client, err := ethclient.Dial(config.BSCChain.WSURL)
	if err != nil {
		log.LogFatal("BSC WebSocket 连接失败: %v", err)
	}
	defer client.Close()

	// 将客户端设置为全局，便于其他包使用（event、trade 等）
	manager.SetClient(client)

	log.LogInfo("成功连接 BSC WebSocket: %s", config.BSCChain.WSURL)

	// 6. 启动所有监听器
	startAllListeners()

	// 7. 启动持仓盈亏监控协程
	manager.WG.Add(1)
	go monitorProfitAndLoss()

	// 8. 程序主循环，保持运行
	log.LogInfo("Fourmeme 监听 & 自动交易机器人已启动（当前时间: %s）", time.Now().Format("2006-01-02 15:04:05"))
	log.LogInfo("监控持仓、止盈止损自动执行中... 按 Ctrl+C 退出")

	// 阻塞主 goroutine
	select {}
}

// startAllListeners 启动所有配置的监听任务
func startAllListeners() {
	// 主监听目标：Fourmeme Token + Manager 合约事件
	for _, target := range config.DefaultMonitorTargets {
		manager.WG.Add(1)
		go func(t *config.MonitorTarget) {
			defer manager.WG.Done()
			event.StartTokenListener(manager.GetClient(), t)
		}(target)
	}

	// 聪明钱包监听（可选）
	if config.DefaultSmartWallets.Enabled && len(config.DefaultSmartWallets.Wallets) > 0 {
		manager.WG.Add(1)
		go func() {
			defer manager.WG.Done()
			event.StartSmartWalletListener(manager.GetClient(), config.DefaultSmartWallets)
		}()
	}

	// Creator 监听（可选）
	if config.DefaultCreators.Enabled && len(config.DefaultCreators.Creators) > 0 {
		manager.WG.Add(1)
		go func() {
			defer manager.WG.Done()
			event.StartCreatorListener(manager.GetClient(), config.DefaultCreators)
		}()
	}

	log.LogInfo("所有监听器已启动 → Token目标: %d | 聪明钱包: %v | Creator: %v",
		len(config.DefaultMonitorTargets),
		config.DefaultSmartWallets.Enabled,
		config.DefaultCreators.Enabled)
}

// monitorProfitAndLoss 定期检查所有持仓，实现止盈止损自动卖出
func monitorProfitAndLoss() {
	defer manager.WG.Done()

	ticker := time.NewTicker(10 * time.Second) // 每10秒检查一次
	defer ticker.Stop()

	log.LogInfo("持仓盈亏监控协程已启动（检查间隔: 10s）")

	for {
		select {
		case <-ticker.C:
			positions := manager.GetAllPositions()

			if len(positions) == 0 {
				continue
			}

			for tokenAddr, pos := range positions {
				if pos.Sold {
					continue
				}

				// 获取当前二级市场价格（PancakeSwap）
				currentPriceBNBPerToken, err := trade.GetCurrentTokenPrice(manager.GetClient(), tokenAddr)
				if err != nil {
					log.LogError("获取价格失败 [%s]: %v", tokenAddr[:10], err)
					continue
				}

				// 当前持仓总价值（BNB）
				currentValue := new(big.Float).Mul(new(big.Float).SetInt(pos.BuyTokenAmount), currentPriceBNBPerToken)

				// 盈亏倍数 = 当前价值 / 投入成本
				profitRatio := new(big.Float).Quo(currentValue, pos.BuyAmountBNB)

				profitRatioFloat, _ := profitRatio.Float64()

				log.LogInfo("持仓监控 [%s] | 成本: %.6f BNB | 当前价值: %.6f BNB | 盈亏倍数: %.3fx",
					tokenAddr[:10],
					pos.BuyAmountBNB.Text('f', 6),
					currentValue.Text('f', 6),
					profitRatioFloat)

				// 止盈触发
				if profitRatioFloat >= pos.TargetProfitMult {
					log.LogInfo("【止盈触发】%.2fx → 准备卖出 %s", pos.TargetProfitMult, tokenAddr[:10])
					go executeSell(tokenAddr, pos, "take_profit")
					continue
				}

				// 止损触发
				if profitRatioFloat <= pos.TargetLossMult {
					log.LogInfo("【止损触发】%.2fx → 准备卖出 %s", pos.TargetLossMult, tokenAddr[:10])
					go executeSell(tokenAddr, pos, "stop_loss")
				}
			}

		case <-manager.ShutdownChan:
			log.LogInfo("收到关闭信号，停止盈亏监控")
			return
		}
	}
}

// executeSell 执行卖出并标记已卖
func executeSell(tokenAddr string, pos *entity.Position, reason string) {
	txHash, err := trade.SellTokenSecondary(manager.GetClient(), tokenAddr, pos.BuyTokenAmount, 0.15) // 15% 滑点
	if err != nil {
		log.LogError("自动卖出失败 [%s] 原因: %s | 错误: %v", tokenAddr[:10], reason, err)
		return
	}

	log.LogInfo("【自动卖出成功】%s | Token: %s | Tx: %s", reason, tokenAddr[:10], txHash)

	// 标记为已卖出
	manager.MarkAsSold(tokenAddr)

	// 记录数据库
	database.SaveTxRecord(&po.TransactionRecord{
		Type:      "sell_" + reason,
		TxHash:    txHash,
		TokenAddr: tokenAddr,
		AmountOut: pos.BuyTokenAmount,
		Status:    "success",
		Timestamp: time.Now(),
	})
}
