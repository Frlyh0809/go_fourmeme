package config

import "math/big"

type ChainConfig struct {
	Name               string   // e.g., "BSC Mainnet"
	RPCURL             string   // HTTP RPC，用于查询和发送交易
	WSURL              string   // WebSocket URL，用于实时订阅事件
	ChainID            *big.Int // BSC Mainnet: 56, Testnet: 97
	GasPriceMultiplier float64  // Gas价格倍数，例如 1.2 表示建议GasPrice * 1.2
	GasLimit           uint64   // 默认Gas限制，可被具体交易覆盖
	MaxGasPrice        *big.Int // 最高可接受Gas价格（Wei）
	PrivateKey         string   // 钱包私钥（用于签名交易，建议从环境变量加载）
	WalletAddress      string   // 对应的公钥地址（可从私钥派生）
	DBDSN              string
}

var BSCChain = &ChainConfig{
	Name:               "BSC Mainnet",
	RPCURL:             "https://bsc-dataseed.binance.org/",
	WSURL:              "wss://bsc-ws-node.nariox.org:443",
	ChainID:            big.NewInt(56),
	GasPriceMultiplier: 1.1,
	GasLimit:           500000,
	MaxGasPrice:        big.NewInt(10000000000), // 10 Gwei
	// PrivateKey 和 WalletAddress 在 loader.go 中加载
}
