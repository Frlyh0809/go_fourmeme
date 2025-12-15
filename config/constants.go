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
