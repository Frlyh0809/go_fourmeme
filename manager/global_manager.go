// manager/global_manager.go
package manager

import (
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
)

// 全局状态变量
var (
	Client       *ethclient.Client
	WG           sync.WaitGroup
	ShutdownChan = make(chan struct{})
)

// SetClient 设置全局客户端（main 中调用一次）
func SetClient(c *ethclient.Client) {
	Client = c
}

// GetClient 获取全局客户端（其他包安全调用）
func GetClient() *ethclient.Client {
	if Client == nil {
		panic("ethclient 未初始化，请先调用 SetClient")
	}
	return Client
}

// CloseShutdown 优雅关闭（可选扩展）
func CloseShutdown() {
	close(ShutdownChan)
}
