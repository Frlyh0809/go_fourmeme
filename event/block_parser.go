// event/listener.go
package event

import (
	"context"
	"math/big"
	"strings"
	"sync"
	"time"

	"go_fourmeme/config"
	configentity "go_fourmeme/entity/config"
	"go_fourmeme/log"
	"go_fourmeme/manager"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	lastProcessedBlock uint64 // 记录上次处理区块号
	pollingMu          sync.Mutex
)

// StartBlockPolling 主流程：区块轮询监听（保证不漏区块）
func StartBlockPolling(interval time.Duration) {
	defer manager.WG.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.LogInfo("区块轮询启动 (间隔: %v)", interval)

	for {
		select {
		case <-ticker.C:
			pollNewBlocks()
		case <-manager.ShutdownChan:
			return
		}
	}
}

// pollNewBlocks 轮询新区块，并发处理
func pollNewBlocks() {
	client := manager.GetClient()

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.LogError("获取最新区块失败: %v", err)
		return
	}

	latestBlock := header.Number.Uint64()

	if lastProcessedBlock == 0 {
		lastProcessedBlock = latestBlock - 10
	}

	if latestBlock <= lastProcessedBlock {
		return
	}

	log.LogInfo("检测到新区块: %d ~ %d", lastProcessedBlock+1, latestBlock)

	var wg sync.WaitGroup
	for blockNum := lastProcessedBlock + 1; blockNum <= latestBlock; blockNum++ {
		wg.Add(1)
		go func(num uint64) {
			defer wg.Done()
			processBlockReceipts(big.NewInt(int64(num)))
		}(blockNum)
	}

	wg.Wait()

	pollingMu.Lock()
	lastProcessedBlock = latestBlock
	pollingMu.Unlock()
}

// processBlockReceipts 使用 eth_getBlockReceipts 获取整个区块所有 Receipt
func processBlockReceipts(blockNum *big.Int) {
	client := manager.GetClient()

	// 构造 BlockNumberOrHash
	blockNumber := rpc.BlockNumber(blockNum.Int64())
	blockNrOrHash := rpc.BlockNumberOrHash{
		BlockNumber: &blockNumber,
	}

	receipts, err := client.BlockReceipts(context.Background(), blockNrOrHash)
	if err != nil {
		log.LogError("eth_getBlockReceipts 失败 (区块 %d): %v", blockNum, err)
		return
	}
	log.LogInfo("区块 %d | hash size: %d ", blockNum, len(receipts))
	// 收集所有日志
	for _, receipt := range receipts {
		if receipt != nil {
			target := config.DefaultMonitorTargets[0]
			var hashLogs []types.Log
			for _, logPtr := range receipt.Logs {
				if logPtr != nil {
					hashLogs = append(hashLogs, *logPtr) // 解引用指针
				}
			}
			if len(hashLogs) > 0 {
				//log.LogInfo("区块 %d | hash: %s | logs size: %v", blockNum, receipt.TxHash.Hex(), len(hashLogs))
				HandleEventV2(hashLogs, receipt, target)
			}
		}
	}

	//for _, vLog := range allLogs {
	//	target := findTargetByAddress(vLog.Address.Hex())
	//	HandleEventV2(vLog, allLogs, target)
	//}
	// 手动过滤
	//filteredLogs := filterLogs(allLogs)

	//log.LogInfo("区块 %d | 总日志: %d | 匹配日志: %d", blockNum, len(allLogs), len(filteredLogs))

	//for _, vLog := range filteredLogs {
	//	target := findTargetByAddress(vLog.Address.Hex())
	//	HandleEvent(vLog, target)
	//}
}

// filterLogs 手动过滤日志（Addresses + Topic0）
func filterLogs(allLogs []types.Log) []types.Log {
	filtered := make([]types.Log, 0, len(allLogs))

	addrMap := getAllAddressesMap()
	topicMap := getAllTopic0Map()

	for _, l := range allLogs {
		if len(l.Topics) == 0 {
			continue
		}

		if _, ok := addrMap[l.Address]; !ok {
			continue
		}

		if _, ok := topicMap[l.Topics[0]]; !ok {
			continue
		}

		filtered = append(filtered, l)
	}

	return filtered
}

// getAllAddressesMap Addresses map（O(1) 查找）
func getAllAddressesMap() map[common.Address]struct{} {
	m := make(map[common.Address]struct{})
	for _, target := range config.DefaultMonitorTargets {
		if target.TokenAddress != "" {
			m[common.HexToAddress(target.TokenAddress)] = struct{}{}
		}
		for _, mgr := range target.FourmemeManagers {
			m[common.HexToAddress(mgr)] = struct{}{}
		}
	}
	m[config.PancakeFactory] = struct{}{}
	m[config.PancakeRouter] = struct{}{}
	return m
}

// getAllTopic0Map Topic0 map（O(1) 查找）
func getAllTopic0Map() map[common.Hash]struct{} {
	m := make(map[common.Hash]struct{})
	for _, target := range config.DefaultMonitorTargets {
		for _, t := range append(append(target.ERC20Topics, target.PancakeTopics...), target.FourmemeTopics...) {
			m[common.HexToHash(t)] = struct{}{}
		}
	}
	return m
}

// findTargetByAddress 根据日志地址匹配 MonitorTarget 配置
func findTargetByAddress(addr string) *configentity.MonitorTarget {
	for _, target := range config.DefaultMonitorTargets {
		if target.TokenAddress == addr {
			return target
		}
		for _, mgr := range target.FourmemeManagers {
			if strings.EqualFold(mgr, addr) {
				return target
			}
		}
	}
	return nil
}
