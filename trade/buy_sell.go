// trade/buy_sell_v1.go
package trade

import (
	"fmt"
	"math/big"

	configentity "go_fourmeme/entity/config"
	//"go_fourmeme/log"
	//"go_fourmeme/manager"
)

// BuyToken 统一买入入口（路由到一级/二级）
func BuyToken(target *configentity.MonitorTarget, tokenAddr string, marketType string) (string, error) {
	switch marketType {
	case "primary":
		return BuyTokenViaManager(target, tokenAddr)
	case "secondary":
		return BuyTokenSecondary(target, tokenAddr)
	default:
		return "", fmt.Errorf("未知市场类型: %s", marketType)
	}
}

// SellToken 统一卖出入口
func SellToken(tokenAddr string, amount *big.Int, slippage float64, marketType string) (string, error) {
	switch marketType {
	case "primary":
		return SellTokenViaManager(tokenAddr, amount, slippage)
	case "secondary":
		return SellTokenSecondary(tokenAddr, amount, slippage)
	default:
		return "", fmt.Errorf("未知市场类型: %s", marketType)
	}
}

// GetCurrentTokenPrice 价格查询 (二级市场)
func GetCurrentTokenPrice(tokenAddr string) (*big.Float, error) {
	// 调用 secondary.go 中的实现
	//return getSecondaryPrice(tokenAddr) // 内部路由
	return nil, nil
}
