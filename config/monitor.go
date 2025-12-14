// config/monitor.go
package config

import "math/big"

type MonitorTarget struct {
	TokenName    string // 代币名称，仅用于日志标识
	TokenAddress string // Fourmem Token 合约地址
	PairAddress  string // 可选：对应的PancakeSwap交易对地址（检测到后动态填充）

	// 需要监听的关键事件 Topic（事件签名 keccak256 hash 的前32字节 hex）
	TopicsToMonitor []string

	// 可选：需要监听的函数 Method ID（前4字节），用于过滤交易输入数据
	MethodIDsToMonitor []string

	// 交易策略参数
	BuyOnLiquidityAdd  bool       // 检测到添加流动性后立即买入
	BuyAmountBNB       *big.Float // 用多少BNB买入（可转为Wei）
	SlippageTolerance  float64    // 滑点容忍度，例如 0.10 表示10%
	SellPercent        float64    // 盈利后卖出比例，例如 0.5 表示卖一半
	TakeProfitMultiple float64    // 几倍盈利卖出，例如 3.0
	StopLossMultiple   float64    // 止损倍数，例如 0.5

	// 新增：联动控制（解耦，可选）
	TriggerOnSmartWalletBuy bool // 只在聪明钱包买入时触发自己的买操作
	TriggerOnCreatorAction  bool // 只在creator创建/操作时触发

	FourmemeManagers []string // Fourmeme Manager合约地址列表

}

// 更新DefaultTopics：添加从交易提取的Fourmeme事件Topic
var DefaultTopics = []string{
	"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", // Transfer
	"0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925", // Approval
	"0x3d0ce9bfc3ed7d6862dbb28b2dea94561fe714a1b4d019aa8af39730d1ad7c3d", // SafeReceived (BNB接收，流动性相关)
	"0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19", // Fourmeme Custom1 (token deposit/confirm)
	"0x741ffc4605df23259462547defeab4f6e755bdc5fbb6d0820727d6d3400c7e0d", // Fourmeme Custom2 (signature/relay confirm)
	// Pancake事件（不变）
	"0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c", // PairCreated
	"0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9", // Mint (Add Liquidity)
	"0x7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65", // Burn (Remove Liquidity)
	"0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822", // Swap
}

// 默认监听的 Method ID（函数选择器前4字节）
var DefaultMethodIDs = []string{
	"0x38ed1739", // swapExactTokensForTokens
	"0x8803dbee", // swapTokensForExactTokens
	"0xf305d719", // addLiquidityETH
	"0xe8e33700", // addLiquidity
	"0xbaa2abde", // removeLiquidityETH
}

// 示例监听目标（实际使用时可加载多个）
var DefaultMonitorTargets = []*MonitorTarget{
	{
		TokenName:          "FourmemeManager",
		FourmemeManagers:   []string{"0x5c952063c7fc8610FFDB798152D69F0B9550762b", "0xF251F83e40a78868FcfA3FA4599Dad6494E46034", "0xEC4549caDcE5DA21Df6E6422d448034B5233bFbC"},
		TokenAddress:       "",
		TopicsToMonitor:    DefaultTopics,
		MethodIDsToMonitor: DefaultMethodIDs,
		BuyOnLiquidityAdd:  true,
		BuyAmountBNB:       big.NewFloat(0.1), // 0.1 BNB
		SlippageTolerance:  0.12,              // 12% 滑点（meme币常用较高）
		SellPercent:        0.6,
		TakeProfitMultiple: 4.0,
		StopLossMultiple:   0.4,
	},
}
