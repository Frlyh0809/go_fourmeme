// entity/position.go
package entity

import (
	"math/big"
	"time"
)

// Position 持仓实体（用于盈亏监控）
type Position struct {
	TokenAddr        string     `json:"token_addr"`
	BuyTxHash        string     `json:"buy_tx_hash"`
	BuyAmountBNB     *big.Float `json:"buy_amount_bnb"`   // 投入 BNB 数量
	BuyTokenAmount   *big.Int   `json:"buy_token_amount"` // 买入 token 数量
	BuyPriceAvg      *big.Float `json:"buy_price_avg"`    // 平均买入价格 (BNB/token)
	BuyTime          time.Time  `json:"buy_time"`
	TargetProfitMult float64    `json:"target_profit_mult"` // 止盈倍数
	TargetLossMult   float64    `json:"target_loss_mult"`   // 止损倍数
	Sold             bool       `json:"sold"`
}
