// client/eth_client.go
package client

import (
	"context"
	"fmt"
	"time"

	"go_fourmeme/config"
	"go_fourmeme/log"

	"github.com/ethereum/go-ethereum/ethclient"
)

// NewEthClient 创建 BSC 客户端
// 支持 WebSocket（推荐，用于实时订阅）和 HTTP fallback
func NewEthClient() (*ethclient.Client, error) {
	var cli *ethclient.Client
	var err error

	// 优先使用 WebSocket（实时事件订阅必需）
	if config.BSCChain.WSURL != "" {
		log.LogInfo("尝试连接 BSC WebSocket: %s", config.BSCChain.WSURL)
		cli, err = ethclient.Dial(config.BSCChain.WSURL)
		if err == nil {
			log.LogInfo("WebSocket 连接成功")
			return cli, nil
		}
		log.LogWarn("WebSocket 连接失败: %v，将尝试 HTTP", err)
	}

	// Fallback 到 HTTP RPC
	if config.BSCChain.RPCURL != "" {
		log.LogInfo("尝试连接 BSC HTTP RPC: %s", config.BSCChain.RPCURL)
		cli, err = ethclient.Dial(config.BSCChain.RPCURL)
		if err != nil {
			log.LogError("HTTP RPC 连接失败: %v", err)
			return nil, err
		}
		log.LogInfo("HTTP RPC 连接成功")
		return cli, nil
	}

	return nil, fmt.Errorf("无可用节点 URL 配置")
}

// NewEthClientWithRetry 带重试的客户端创建（推荐用于生产）
func NewEthClientWithRetry(maxRetries int, retryInterval time.Duration) (*ethclient.Client, error) {
	for i := 0; i <= maxRetries; i++ {
		cli, err := NewEthClient()
		if err == nil {
			return cli, nil
		}
		if i < maxRetries {
			log.LogWarn("客户端连接失败 (尝试 %d/%d): %v，%v 后重试", i+1, maxRetries, err, retryInterval)
			time.Sleep(retryInterval)
		}
	}
	return nil, fmt.Errorf("客户端连接失败，已达最大重试次数")
}

// Ping 测试客户端连通性
func Ping(cli *ethclient.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := cli.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("Ping 失败: %v", err)
	}
	return nil
}
