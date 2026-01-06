package po

import "time"

type WhitelistCreator struct {
	CreatorAddress string    `gorm:"primaryKey;size:128"`
	CreateTime     time.Time `gorm:"index"`
	Reason         string    `gorm:"size:255"`
}
