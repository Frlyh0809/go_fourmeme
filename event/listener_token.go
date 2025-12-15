package event

import (
	"context"
	"go_fourmeme/config"
	"go_fourmeme/log"
	"go_fourmeme/manager"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	// 特定 token 订阅管理（备用）
	tokenSubsMu sync.Mutex
	tokenSubs   = make(map[string]chan<- types.Log)
)

//var transferTopic = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

// 入后启动特定 token 的实时订阅（Transfer 监控）
func StartSpecificTokenSubscription(tokenAddr string) {
	tokenSubsMu.Lock()
	if _, exists := tokenSubs[tokenAddr]; exists {
		tokenSubsMu.Unlock()
		return
	}

	logsChan := make(chan types.Log, 100)
	tokenSubs[tokenAddr] = logsChan
	tokenSubsMu.Unlock()

	client := manager.GetClient()

	transferTopic := common.HexToHash(config.TransferTopic)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(tokenAddr)},
		Topics: [][]common.Hash{
			{transferTopic},
		},
	}

	sub, err := client.SubscribeFilterLogs(context.Background(), query, logsChan)
	if err != nil {
		log.LogError("特定 token %s 订阅失败: %v", tokenAddr[:10], err)
		return
	}

	manager.WG.Add(1)
	go func() {
		defer manager.WG.Done()
		for {
			select {
			case err := <-sub.Err():
				log.LogError("特定 token %s 订阅错误，重连: %v", tokenAddr[:10], err)
				time.Sleep(5 * time.Second)
				StartSpecificTokenSubscription(tokenAddr)
				return
			case vLog := <-logsChan:
				target := findTargetByAddress(tokenAddr)
				HandleEvent(vLog, target) // 处理 Transfer（例如监控卖出信号）
			case <-manager.ShutdownChan:
				sub.Unsubscribe()
				return
			}
		}
	}()

	log.LogInfo("特定 token %s 实时订阅启动（监控 Transfer）", tokenAddr[:10])
}
