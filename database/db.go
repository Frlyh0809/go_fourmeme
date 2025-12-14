// database/db.go
package database

import (
	"go_fourmeme/config"
	"go_fourmeme/entity/po"
	"go_fourmeme/log"
	"gorm.io/driver/sqlite" // 或其他驱动，如 postgres
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化数据库连接并迁移表
func InitDB(cfg *config.ChainConfig) {
	dsn := cfg.DBDSN
	if dsn == "" {
		dsn = "transactions.db" // 默认 SQLite 文件
		log.LogInfo("使用默认 SQLite 数据库: %s", dsn)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: nil, // 可自定义日志
	})
	if err != nil {
		log.LogFatal("数据库连接失败: %v", err)
	}

	// 自动迁移所有表（从 entity/po 导入）
	err = DB.AutoMigrate(&po.TransactionRecord{})
	if err != nil {
		log.LogFatal("表迁移失败: %v", err)
	}

	log.LogInfo("数据库初始化成功（DSN: %s）", dsn)
}
