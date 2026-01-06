// entity/po/transaction_record.go
package po

import (
	"time"
)

// TransactionRecord 数据库交易记录实体
type TransactionRecord struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	TxHash    string    `gorm:"uniqueIndex;size:66"`
	Type      string    `gorm:"index;size:64"` // buy_manager, sell_take_profit 等
	TokenAddr string    `gorm:"index;size:42"`
	AmountIn  string    `gorm:"type:varchar(100)"` // 输入金额 (Wei)
	AmountOut string    `gorm:"type:varchar(100)"` // 输出金额 (Wei)
	Slippage  float64   `gorm:"type:decimal(10,4)"`
	Status    string    `gorm:"size:16"` // success, failed, pending
	ErrorMsg  string    `gorm:"type:text"`
	Timestamp time.Time `gorm:"index"`
}
