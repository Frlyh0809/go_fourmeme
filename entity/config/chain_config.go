// entity/config/chain_config.go
package config

import "math/big"

// ChainConfig 链相关全局配置
type ChainConfig struct {
	Name               string   `json:"name" yaml:"name"`
	RPCURL             string   `json:"rpc_url" yaml:"rpc_url"`
	WSURL              string   `json:"ws_url" yaml:"ws_url"`
	ChainID            *big.Int `json:"chain_id" yaml:"chain_id"`
	GasPriceMultiplier float64  `json:"gas_price_multiplier" yaml:"gas_price_multiplier"`
	GasLimit           uint64   `json:"gas_limit" yaml:"gas_limit"`
	MaxGasPrice        *big.Int `json:"max_gas_price" yaml:"max_gas_price"`

	PrivateKey    string `json:"-" yaml:"-"`                           // 从 env 加载，不写文件
	WalletAddress string `json:"wallet_address" yaml:"wallet_address"` // 自动派生

	DBDSN string `json:"db_dsn" yaml:"db_dsn"` // 数据库连接串
}
