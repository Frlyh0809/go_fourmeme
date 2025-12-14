// manager/global_manager.go
package manager

import (
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	Client       *ethclient.Client
	WG           sync.WaitGroup
	ShutdownChan = make(chan struct{})
)

// SetClient 设置全局客户端（main 中调用）
func SetClient(c *ethclient.Client) {
	Client = c
}

// GetClient 获取客户端
func GetClient() *ethclient.Client {
	return Client
}
