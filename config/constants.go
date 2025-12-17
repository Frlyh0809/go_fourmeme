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
	USDTAddress        = "0x55d398326f99059fF775485246999027B3197955" //BUSDT
	USD1Address        = "0x8d0d000ee44948fc98c9b98a4fa4921476f08b0d" //WLFI: USD1
	ZeroAddress        = "0x0000000000000000000000000000000000000000"
)

const (
	TopicTransfer = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

	TopicManager2OwnershipTransferred = "0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0"
	TopicManager2CreateEvent1         = "0x396d5e902b675b032348d3d2e9517ee8f0c4a926603fbc075d3d282ff00cad20"

	// Manager1 旧事件 (很少用，但兼容)
	TopicManager1Event1 = "0x80d4e495cda89b31af98c8e977ff11f417bafcee26902a17a15be51830c47533"
	TopicManager1Event2 = "0x00fe0e12b43090c1fc19a34aefa5cc138a4eeafc60ab800f855c730b3fb9480e"

	// Manager2 / Helper3 新事件
	TopicManager2Event1 = "0x7db52723a3b2cdd6164364b3b766e65e540d7be48ffa89582956d8eaebe62942"
	TopicManager2Event2 = "0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19"

	//PancakeSwap
	PairCreatedTopic    = "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c"
	MintTopic           = "0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9"
	DepositConfirmTopic = "0x0a5575b3648bae2210cee56bf33254cc1ddfbc7bf637c0af2ac18b14fb1bae19"
)

// 方便使用的 Address 变量
var (
	DefaultFourmemeManagerAddr = common.HexToAddress(DefaultFourmemeManager)
	AddrTokenManagerHelper3    = common.HexToAddress(TokenManagerHelper3)
	AddrTokenManager1          = common.HexToAddress(TokenManager1)
	PancakeFactory             = common.HexToAddress(PancakeFactoryAddr)
	PancakeRouter              = common.HexToAddress(PancakeRouterAddr)
	WBNB                       = common.HexToAddress(WBNBAddr)

	HashTransfer                  = common.HexToHash(TopicTransfer)
	HashOwnershipTransferred      = common.HexToHash(TopicManager2OwnershipTransferred)
	HashTopicManager2CreateEvent1 = common.HexToHash(TopicManager2CreateEvent1)
	HashManager1Event1            = common.HexToHash(TopicManager1Event1)
	HashManager1Event2            = common.HexToHash(TopicManager1Event2)
	HashManager2Event1            = common.HexToHash(TopicManager2Event1)
	HashManager2Event2            = common.HexToHash(TopicManager2Event2)

	AddrUSDT = common.HexToAddress(USDTAddress)
	AddrUSD1 = common.HexToAddress(USD1Address)
	AddrZero = common.HexToAddress(ZeroAddress)
)
