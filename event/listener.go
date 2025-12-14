// event/listener.go
package event

import (
	"context"
	"sync"

	"go_fourmeme/config"
	"go_fourmeme/log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	mu sync.Mutex // 用于保护动态监听
)

// StartTokenListener: 为单个MonitorTarget启动监听（token + Fourmeme Managers + Pancake相关）
func StartTokenListener(client *ethclient.Client, target *config.MonitorTarget) {
	if target == nil {
		return
	}

	var addresses []common.Address

	// 监听特定token（如果指定）
	if target.TokenAddress != "" {
		addresses = append(addresses, common.HexToAddress(target.TokenAddress))
	}

	// 监听Fourmeme Manager合约（核心）
	for _, mgr := range target.FourmemeManagers {
		addresses = append(addresses, common.HexToAddress(mgr))
	}

	// 监听PancakeSwap Factory/Router（流动性事件）
	pancakeFactory := common.HexToAddress(config.PancakeFactory) // Pancake V2 Factory
	pancakeRouter := common.HexToAddress(config.PancakeRouter)   // Pancake V2 Router
	addresses = append(addresses, pancakeFactory, pancakeRouter)

	// Topics过滤
	var topics [][]common.Hash
	for _, t := range target.TopicsToMonitor {
		topics = append(topics, []common.Hash{common.HexToHash(t)})
	}

	query := ethereum.FilterQuery{
		Addresses: addresses,
		Topics:    topics,
	}

	logs := make(chan types.Log, 100)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.LogFailure(err, "订阅失败")
		return
	}

	log.LogInfo("启动监听目标: %v, 地址数: %d", target.TokenName, len(addresses))

	for {
		select {
		case err := <-sub.Err():
			log.LogFailure(err, "订阅错误，重连")
			// 重连逻辑
			StartTokenListener(client, target)
			return
		case vLog := <-logs:
			go HandleEvent(vLog, target, client) // 并发处理事件，避免阻塞
			//go HandleEvent(vLog, target, client) // 并发处理事件，避免阻塞
		}
	}
}

// StartSmartWalletListener: 监听聪明钱包（独立订阅）
func StartSmartWalletListener(client *ethclient.Client, cfg *config.SmartWalletsConfig) {
	if !cfg.Enabled || len(cfg.Wallets) == 0 {
		return
	}

	var addresses []common.Address
	for _, w := range cfg.Wallets {
		addresses = append(addresses, common.HexToAddress(w.WalletAddress))
	}

	// 使用默认topics或自定义
	var topics [][]common.Hash
	for _, t := range config.DefaultTopics {
		topics = append(topics, []common.Hash{common.HexToHash(t)})
	}

	query := ethereum.FilterQuery{
		Addresses: addresses,
		Topics:    topics,
	}

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.LogFailure(err, "聪明钱包订阅失败")
	}

	log.LogInfo("启动聪明钱包监听: %d 个地址", len(addresses))

	for {
		select {
		case err := <-sub.Err():
			log.LogFailure(err, "未知")
			StartSmartWalletListener(client, cfg)
			return
		case vLog := <-logs:
			go HandleEvent(vLog, nil, client) // target为nil，表示聪明钱包事件
			//go HandleEvent(vLog) // target为nil，表示聪明钱包事件
		}
	}
}

// StartCreatorListener: 类似聪明钱包，监听creator地址
func StartCreatorListener(client *ethclient.Client, cfg *config.CreatorsConfig) {
	if !cfg.Enabled || len(cfg.Creators) == 0 {
		return
	}

	// 实现类似StartSmartWalletListener，略（根据需求复制修改）
}
