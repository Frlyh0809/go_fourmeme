// entity/po/transaction_create.go
package po

import (
	"time"
)

type TransactionCreate struct {
	TxUniqueSeq     string    `gorm:"primaryKey;size:30;comment:'block + txIndex+logIndex'"`
	BlockNumber     string    `gorm:"type:varchar(30)"`
	TxIndex         string    `gorm:"size:16"`
	TxHash          string    `gorm:"size:128;index"`
	Send            string    `gorm:"size:128;index;comment:'交易发起人'"`
	Receive         string    `gorm:"size:128;index;comment:'接受人'"`
	Type            int       `gorm:"comment:'0:创建 6出盘 7转移owner 8销毁'"`
	Protocol        string    `gorm:"size:128;comment:'协议[managerV1,managerV2,managerV3]'"`
	ProtocolAddress string    `gorm:"size:128;index"`
	TokenAddress    string    `gorm:"size:128;index;comment:'交易的token地址'"`
	PaymentToken    string    `gorm:"size:128;comment:'交易支付的代币'"`
	TokenAmount     string    `gorm:"type:varchar(100)"`
	PaymentAmount   string    `gorm:"type:varchar(100)"`
	Price           string    `gorm:"type:varchar(100)"`
	Volume          string    `gorm:"type:varchar(100)"`
	CreatedAt       time.Time `gorm:"index"`
	UpdatedAt       time.Time `gorm:"index"`
}

func (TransactionCreate) TableName() string {
	return "bsc_fourmeme_transaction_create"
}
