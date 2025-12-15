// database/db.go
package database

import (
	"go_fourmeme/config"
	"go_fourmeme/entity/po" // 数据库实体
	"go_fourmeme/log"

	"gorm.io/driver/sqlite" // 可替换为 postgres/mysql
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化数据库连接并迁移表
func InitDB() {
	dsn := config.BSCChain.DBDSN
	if dsn == "" {
		dsn = "fourmeme_transactions.db" // 默认本地 SQLite
		log.LogInfo("使用默认 SQLite 数据库: %s", dsn)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.LogFatal("数据库连接失败: %v", err)
	}

	// 自动迁移实体表
	err = DB.AutoMigrate(&po.TransactionRecord{})
	if err != nil {
		log.LogFatal("数据库表迁移失败: %v", err)
	}

	log.LogInfo("数据库初始化成功 (DSN: %s)", dsn)
}

// CloseDB 优雅关闭（main defer 调用，可选）
func CloseDB() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		sqlDB.Close()
	}
}
