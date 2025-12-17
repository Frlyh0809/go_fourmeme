// manager/global_manager.go
package manager

import (
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
)

// 全局状态变量
var (
	EthClient    *ethclient.Client
	WG           sync.WaitGroup
	ShutdownChan = make(chan struct{})
)

// SetEthClient 设置全局客户端（main 中调用一次）
func SetEthClient(c *ethclient.Client) {
	EthClient = c
}

// GetEthClient 获取全局客户端（其他包安全调用）
func GetEthClient() *ethclient.Client {
	if EthClient == nil {
		panic("ethclient 未初始化，请先调用 SetEthClient")
	}
	return EthClient
}

// CloseShutdown 优雅关闭（可选扩展）
func CloseShutdown() {
	close(ShutdownChan)
}
