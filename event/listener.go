// event/listener.go
package event

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go_fourmeme/config"
	configentity "go_fourmeme/entity/config" // MonitorTarget 实体
	"go_fourmeme/log"
	"go_fourmeme/manager"
)

var listenerMu sync.Mutex // 防止重复订阅

// StartTokenListener 为单个 MonitorTarget 启动监听
func StartTokenListener(target *configentity.MonitorTarget) {
	if target == nil {
		return
	}

	client := manager.GetEthClient()

	var addresses []common.Address

	// Token 地址（如果指定）
	if target.TokenAddress != "" {
		addresses = append(addresses, common.HexToAddress(target.TokenAddress))
	}

	// Fourmeme Manager 地址
	for _, mgr := range target.FourmemeManagers {
		addresses = append(addresses, common.HexToAddress(mgr))
	}

	// PancakeSwap 核心地址（固定）
	addresses = append(addresses, config.PancakeFactory, config.PancakeRouter)

	// 所有 Topic 合并（从配置隔离加载）
	var topics [][]common.Hash
	for _, t := range append(append(target.ERC20Topics, target.PancakeTopics...), target.FourmemeTopics...) {
		topics = append(topics, []common.Hash{common.HexToHash(t)})
	}

	// MethodID 过滤可通过 Transaction 过滤，但 FilterQuery 不支持，这里暂不实现（可用 fallback 轮询）

	query := ethereum.FilterQuery{
		Addresses: addresses,
		Topics:    topics,
	}

	logsChan := make(chan types.Log, 100)

	sub, err := client.SubscribeFilterLogs(context.Background(), query, logsChan)
	if err != nil {
		log.LogFatal("订阅失败 [Target: %s]: %v1", target.TokenName, err)
	}

	log.LogInfo("监听启动成功 [Target: %s] | 地址数: %d | Topic数: %d", target.TokenName, len(addresses), len(topics))

	manager.WG.Add(1)
	go func() {
		defer manager.WG.Done()
		for {
			select {
			case err := <-sub.Err():
				log.LogError("订阅错误，重连 [Target: %s]: %v1", target.TokenName, err)
				// 重连逻辑
				time.Sleep(5 * time.Second)
				StartTokenListener(target)
				return
			case vLog := <-logsChan:
				go HandleEvent(vLog, target) // 并发处理，避免阻塞
			case <-manager.ShutdownChan:
				sub.Unsubscribe()
				return
			}
		}
	}()
}

// StartAllListeners 启动所有配置的监听（main 中调用）
func StartAllListeners() {
	for _, target := range config.DefaultMonitorTargets {
		StartTokenListener(target)
	}

	// 聪明钱包和 Creator 监听可类似实现（略）
}
