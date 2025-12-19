package trade

import (
	"go_fourmeme/client" // 只导入必要
	configentity "go_fourmeme/entity/config"
)

// Buy 入口买入路由
func Buy(target *configentity.MonitorTarget, tokenAddr string) (string, error) {
	// 判断一级/二级 (假设基于 client 或 config)
	if isPrimaryMarket(tokenAddr) { // 你实现这个判断
		return primaryBuy(target, tokenAddr)
	}
	// 二级市场逻辑 (PancakeSwap 等)
	//return secondaryBuy(target, tokenAddr)
	return "", nil
}

// Sell 入口卖出路由
func Sell(target *configentity.MonitorTarget, tokenAddr string) (string, error) {
	if isPrimaryMarket(tokenAddr) {
		return primarySell(target, tokenAddr)
	}
	// 二级市场逻辑
	//return secondarySell(target, tokenAddr)
	return "", nil
}

// isPrimaryMarket 判断一级市场 (示例)
func isPrimaryMarket(tokenAddr string) bool {
	info := client.GetTokenStatus(tokenAddr)
	return info.Status == client.StatusTrading // 假设有Type
}
