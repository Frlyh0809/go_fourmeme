package test

import (
	"fmt"
	"go_fourmeme/client"
	"go_fourmeme/utils"

	//"go_fourmeme/client"
	"go_fourmeme/config"
	"go_fourmeme/trade"
	"go_fourmeme/trade/internal_market/v3"
	"math/big"
	"testing"
)

func TestBuyTokenViaManager3(t *testing.T) {
	Before()

	target := config.DefaultMonitorTargets[0]
	tokenAddr := "0x1501223E56F6c18df9dfF8c7f5a183C6207d4444"
	amountInWei := trade.ToWei(target.BuyAmountBNB) // 这就是 funds，通过 Value 传递

	// 从 getTokenInfo 获取当前 funds
	//info, err := client.GetTokenStatus(tokenAddr)
	//if info.Funds == nil {
	//
	//}
	//currentFunds := info.Funds
	//
	//res, err := v3.TryBuy(tokenAddr, amountInWei, currentFunds)
	//if err != nil {
	//	fmt.Println("try buy error", err)
	//}
	//fmt.Println("try buy", res)

	privateKey, err := trade.GetPrivateKey()
	if err != nil {
	}
	hash, err := v3.BuyTokenViaManagerV3(tokenAddr, amountInWei, big.NewInt(1), privateKey)
	fmt.Println("hash", hash)
}

func TestSellTokenViaManager3(t *testing.T) {
	Before()

	//target := config.DefaultMonitorTargets[0]
	tokenAddr := "0x1501223E56F6c18df9dfF8c7f5a183C6207d4444"
	//amountInWei := trade.ToWei(target.BuyAmountBNB) // 这就是 funds，通过 Value 传递

	tokenAmounts := utils.Mul10Pow(big.NewFloat(224530.98718979), big.NewInt(18))
	fmt.Println("tokenAmounts", tokenAmounts)
	//tokenAmounts := big.NewInt(224530)
	// 从 getTokenInfo 获取当前 funds
	info, err := client.GetTokenStatus(tokenAddr)
	if info.Funds == nil {

	}
	fmt.Println("---")
	//currentFunds := info.Funds

	//res, err := v3.TrySell(tokenAddr, tokenAmounts)
	//if err != nil {
	//	fmt.Println("try sell error", err)
	//}
	//fmt.Println("try sell", res)

	privateKey, err := trade.GetPrivateKey()
	if err != nil {
	}

	hash, err := v3.SellTokenViaManagerV3(tokenAddr, tokenAmounts, big.NewInt(0), privateKey)
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("hash", hash)
}
