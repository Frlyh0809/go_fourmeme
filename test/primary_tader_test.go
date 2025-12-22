package test

import (
	"fmt"
	"go_fourmeme/client"
	"go_fourmeme/config"
	"go_fourmeme/trade"
	"testing"
)

func TestPrimaryBuy(t *testing.T) {
	target := config.DefaultMonitorTargets[0]
	tokenAddr := ""
	trade.PrimaryBuy(target, tokenAddr)
}

func TestInfo(t *testing.T) {
	Before()
	{
		//v1
		//0xd3f1F54A500aC8c0abdFc5cCB62d6cA2a714e391
		tokenAddr := "0xd3f1F54A500aC8c0abdFc5cCB62d6cA2a714e391"
		tokenInfo, _ := client.GetTokenStatus(tokenAddr)
		fmt.Println(tokenInfo)
	}
	//{
	//
	//	//Quote 0x55d398326f99059fF775485246999027B3197955 B-USDT
	//	tokenAddr := "0xc825316bfedc7672a013024d5b8230c449744444"
	//	tokenInfo, _ := client.GetTokenStatus(tokenAddr)
	//	fmt.Println(tokenInfo)
	//}
	//println("------")
	//{
	//	//Quote 0x0000000000000000000000000000000000000000 bnb
	//	tokenAddr := "0xd28535379a9a894dc1f1e107e482c7b1b6934444"
	//	tokenInfo, _ := client.GetTokenStatus(tokenAddr)
	//	fmt.Println(tokenInfo)
	//}
	//println("------")
	//{
	//	//Quote 0x8d0D000Ee44948FC98c9B98A4FA4921476f08B0d USD1
	//	tokenAddr := "0x1501223E56F6c18df9dfF8c7f5a183C6207d4444"
	//	tokenInfo, _ := client.GetTokenStatus(tokenAddr)
	//	fmt.Println(tokenInfo)
	//}

}
