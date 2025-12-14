// entity/po/transaction_record.go
package po

import (
	"math/big"
	"time"
)

type TransactionRecord struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"` // 主键
	TxHash    string    `gorm:"uniqueIndex;size:66"`      // 交易Hash (0x... 66字符)
	Type      string    `gorm:"index;size:32"`            // 类型: buy_manager, sell_take_profit 等
	TokenAddr string    `gorm:"index;size:42"`            // Token 地址
	AmountIn  *big.Int  `gorm:"type:varchar(255)"`        // 输入金额 (Wei/BNB)
	AmountOut *big.Int  `gorm:"type:varchar(255)"`        // 输出金额 (Wei/Token)
	Slippage  float64   // 滑点百分比
	Status    string    `gorm:"size:16"`   // success, failed, pending
	ErrorMsg  string    `gorm:"type:text"` // 错误消息
	Timestamp time.Time `gorm:"index"`     // 时间戳
}
