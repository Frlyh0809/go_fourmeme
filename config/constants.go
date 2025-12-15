// config/constants.go
package config

import (
	"github.com/ethereum/go-ethereum/common"
)

// Fourmeme 核心合约地址 (BSC 主网)
const (
	DefaultFourmemeManager = "0x5c952063c7fc8610FFDB798152D69F0B9550762b" // TokenManager2
	TokenManagerHelper3    = "0xF251F83e40a78868FcfA3FA4599Dad6494E46034"
	TokenManager1          = "0xEC4549caDcE5DA21Df6E6422d448034B5233bFbC"
)

// TODO PancakeSwap 一共四个版本 需要去分析 看fourmeme用的哪几套
// PancakeSwap V2 主网地址
const (
	PancakeFactoryAddr = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
	PancakeRouterAddr  = "0x10ED43C718714eb63d5aA57df234750"
	WBNBAddr           = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"
)

// 方便使用的 Address 变量
var (
	DefaultFourmemeManagerAddr = common.HexToAddress(DefaultFourmemeManager)
	PancakeFactory             = common.HexToAddress(PancakeFactoryAddr)
	PancakeRouter              = common.HexToAddress(PancakeRouterAddr)
	WBNB                       = common.HexToAddress(WBNBAddr)
)

const (
	TransferTopic       = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	PairCreatedTopic    = "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c"
	MintTopic           = "0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9"
	DepositConfirmTopic = "0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19"
)
