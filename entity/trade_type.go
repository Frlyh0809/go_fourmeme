package entity

import (
	"math/big"
	"time"
)

type Position struct {
	TokenAddr        string
	BuyTxHash        string
	BuyAmountBNB     *big.Float // 投入BNB数量
	BuyTokenAmount   *big.Int   // 买入token数量
	BuyPriceAvg      *big.Float // 平均买入价格（BNB/token）
	BuyTime          time.Time
	TargetProfitMult float64 // 止盈倍数（如 4.0）
	TargetLossMult   float64 // 止损倍数（如 0.5）
	Sold             bool
}
