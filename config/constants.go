// config/constants.go  或 config/monitor.go
package config

import "github.com/ethereum/go-ethereum/common"

const (
	// Fourmeme 核心合约地址（BSC 主网）
	DefaultFourmemeManager = "0x5c952063c7fc8610FFDB798152D69F0B9550762b" // TokenManager2 主地址
	TokenManagerHelper3    = "0xF251F83e40a78868FcfA3FA4599Dad6494E46034"
	TokenManager1          = "0xEC4549caDcE5DA21Df6E6422d448034B5233bFbC"

	// PancakeSwap V2 主网地址
	PancakeFactory = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
	PancakeRouter  = "0x10ED43C718714eb63d5aA57df234750"
	WBNB           = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"
)

// 方便比较的 Address 版本（可选）
var (
	DefaultFourmemeManagerAddr = common.HexToAddress(DefaultFourmemeManager)
)
