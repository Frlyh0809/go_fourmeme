// config/smart_wallets.go
package config

import "math/big"

type SmartWallet struct {
	WalletAddress      string   // 钱包地址
	TopicsToMonitor    []string // 自定义事件Topic（e.g., 只监听Transfer/Swap）
	MethodIDsToMonitor []string // 自定义Method ID（e.g., swapExactTokensForTokens）

	// 策略：检测到该钱包行为后触发什么
	OnBuyAction        string   // e.g., "buy" - 跟随买入
	OnSellAction       string   // e.g., "sell" - 跟随卖出，或 "ignore"
	MinAmountThreshold *big.Int // 最小交易金额阈值（Wei），低于此忽略
}

type SmartWalletsConfig struct {
	Enabled bool           // 是否启用整个聪明钱包监听（全局开关）
	Wallets []*SmartWallet // 钱包列表，如果为空则不监听
}

// 示例默认配置（可为空）
var DefaultSmartWallets = &SmartWalletsConfig{
	Enabled: true,
	Wallets: []*SmartWallet{
		{
			WalletAddress: "0xSmartWallet1Address",
			TopicsToMonitor: []string{
				"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", // Transfer
				"0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822", // Swap
			},
			MethodIDsToMonitor: []string{"0x38ed1739"}, // swapExactTokensForTokens
			OnBuyAction:        "buy",
			OnSellAction:       "sell",
			MinAmountThreshold: big.NewInt(1000000000000000000), // 1 ETH/BNB 等值
		},
		// 更多钱包...
	},
}
