// main.go
package main

import (
	"go_fourmeme/client"
	"go_fourmeme/entity"
	"go_fourmeme/trade"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go_fourmeme/config"
	"go_fourmeme/database"
	"go_fourmeme/log"
	"go_fourmeme/manager"
	"go_fourmeme/parser"
	"go_fourmeme/utils"
)

func main() {
	// 1. 加载配置
	configFile := "config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	config.LoadFromEnvAndFile(configFile)

	// 2. 初始化日志
	log.InitLogger()

	// 3. 加载 ABI 文件
	if err := utils.LoadABIs(); err != nil {
		log.LogFatal("ABI 加载失败: %v1", err)
	}

	// 4. 初始化数据库
	database.InitDB()
	client.InitBnbPriceCache()
	// 5. 连接客户端并设置全局
	ethClient, err := client.NewEthClientWithRetry(5, 5*time.Second)
	if err != nil {
		log.LogFatal("BSC 客户端连接失败: %v1", err)
	}
	manager.SetEthClient(ethClient)
	defer ethClient.Close()

	// 6. 启动所有监听
	//parser.StartAllListeners()
	// 启动主流程：区块轮询监听
	manager.WG.Add(1)
	go parser.StartBlockPolling(3 * time.Second) // 每 3 秒轮询一次

	// 7. 启动盈亏监控
	manager.WG.Add(1)
	//go monitorProfitAndLoss()

	log.LogInfo("Fourmeme 自动交易机器人已启动！监控中...")

	// 8. 优雅关闭
	waitForShutdown()
}

func monitorProfitAndLoss() {
	defer manager.WG.Done()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			positions := manager.GetAllPositions()
			for tokenAddr, pos := range positions {
				if pos.Sold {
					continue
				}
				//currentPrice, err := trade.GetCurrentTokenPrice(tokenAddr)
				//if err != nil {
				//	continue
				//}
				//TODO 价格
				currentPrice := big.NewFloat(0)

				currentValue := new(big.Float).Mul(new(big.Float).SetInt(pos.BuyTokenAmount), currentPrice)
				ratio := new(big.Float).Quo(currentValue, pos.BuyAmountBNB)

				ratioFloat, _ := ratio.Float64()
				if ratioFloat >= pos.TargetProfitMult {
					go executeSell(tokenAddr, pos, "take_profit")
				} else if ratioFloat <= pos.TargetLossMult {
					go executeSell(tokenAddr, pos, "stop_loss")
				}
			}
		case <-manager.ShutdownChan:
			return
		}
	}
}

func executeSell(tokenAddr string, pos *entity.Position, reason string) {
	txHash, err := trade.SellTokenSecondary(tokenAddr, pos.BuyTokenAmount, 0.15)
	if err != nil {
		log.LogError("卖出失败 [%s] %s: %v1", reason, tokenAddr[:10], err)
		return
	}
	manager.MarkAsSold(tokenAddr)
	log.LogInfo("自动卖出成功 [%s] Token: %s Tx: %s", reason, tokenAddr[:10], txHash)
}

func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.LogInfo("收到关闭信号，正在优雅退出...")
	close(manager.ShutdownChan)
	manager.WG.Wait()
	database.CloseDB()
	log.LogInfo("程序已退出")
}
