// entity/config/smart_wallets.go
package config

import "math/big"

// SmartWallet 单个聪明钱包配置
type SmartWallet struct {
	WalletAddress      string   `json:"wallet_address" yaml:"wallet_address"`
	OnBuyAction        string   `json:"on_buy_action" yaml:"on_buy_action"` // "buy" / "ignore"
	OnSellAction       string   `json:"on_sell_action" yaml:"on_sell_action"`
	MinAmountThreshold *big.Int `json:"min_amount_threshold" yaml:"min_amount_threshold"` // Wei
}

// SmartWalletsConfig 聪明钱包组配置
type SmartWalletsConfig struct {
	Enabled bool           `json:"enabled" yaml:"enabled"`
	Wallets []*SmartWallet `json:"wallets" yaml:"wallets"`
}
