// config/defaults.go
package config

import (
	"math/big"

	"go_fourmeme/entity/config" // 实体结构体路径
)

// DefaultMonitorTargets 默认监听目标（可被配置文件覆盖）
var DefaultMonitorTargets = []*config.MonitorTarget{
	{
		TokenName:        "FourmemeDefault",
		TokenAddress:     "", // 动态发现新 token
		FourmemeManagers: []string{DefaultFourmemeManager, TokenManagerHelper3, TokenManager1},

		// === 事件 Topic 完全隔离 ===
		ERC20Topics: []string{
			TopicTransfer, // Transfer
			//"0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925", // Approval
		},
		//PancakeTopics: []string{
		//	"0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c", // PairCreated
		//	"0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9", // Mint (Add Liquidity)
		//	"0x7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65", // Burn (Remove Liquidity)
		//	"0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822", // Swap
		//},
		FourmemeTopics: []string{
			TopicManager1Event1, // TokenManager1
			TopicManager1Event2, // TokenManager1

			TopicManager2Event1, // TokenManager2 Deposit/Confirm (常见)
			TopicManager2Event2, // TokenManager2 Deposit/Confirm (常见)

			TopicManager2CreateEvent1,         //TokenManager2 createToken
			TopicManager2OwnershipTransferred, //TokenManager2 OwnershipTransferred [previousOwner=零地址：createToken]

			//"0x741ffc4605df23259462547defeab4f6e755bdc5fbb6d0820727d6d3400c7e0d", // Signature/Relay Confirm
			//"0x3d0ce9bfc3ed7d6862dbb28b2dea94561fe714a1b4d019aa8af39730d1ad7c3d", // SafeReceived (BNB) fee
		},

		//MethodIDsToMonitor: []string{
		//	"0xe63aaf36", // 示例 Fourmeme 方法
		//	"0x38ed1739", // swapExactTokensForTokens (Pancake)
		//	"0x5b9e9006", //swapV2ExactIn https://bscscan.com/tx/0x2d204e1e896127af4938a9f85604dd96ac75298ac1e22f80ade8b293db4a03b2
		//	"0xb80c2f09", //smartSwapByOrderId https://bscscan.com/tx/0x3eb37dc9ff17cafcb3ad7c065697dc12d4e1d4e9174d55a6657cd3e4b0b1c5c3
		//},

		// 默认交易策略
		BuyOnLiquidityAdd:       true,
		BuyAmountBNB:            big.NewFloat(0.01), // 0.1 BNB
		SlippageTolerance:       0.12,               // 12%
		TakeProfitMultiple:      4.0,
		StopLossMultiple:        0.5,
		TriggerOnSmartWalletBuy: true,
	},
}

// DefaultSmartWallets / DefaultCreators 可类似定义（略，可复制原有）
var DefaultSmartWallets *config.SmartWalletsConfig
var DefaultCreators *config.CreatorsConfig

var BSCChain = &config.ChainConfig{
	Name:               "BSC Mainnet",
	RPCURL:             "https://api.zan.top/node/v1/bsc/mainnet/79ff3388de8e4280be9e6d32aea48bb9",
	WSURL:              "",
	ChainID:            big.NewInt(56),
	GasPriceMultiplier: 1.1,
	GasLimit:           800000,
	MaxGasPrice:        big.NewInt(10e9), // 10 Gwei
	PrivateKey:         "",               // 从 env 加载
	WalletAddress:      "",
	DBDSN:              "",
}
