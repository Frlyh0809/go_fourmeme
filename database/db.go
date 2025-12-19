// database/db.go
package database

import (
	"go_fourmeme/config"
	"go_fourmeme/log"

	//"gorm.io/driver/sqlite" // 可替换为 postgres/mysql
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化数据库连接并迁移表
func InitDB() {
	dsn := config.BSCChain.DBDSN
	if dsn == "" {
		dsn = "fourmeme_transactions.db" //
		log.LogInfo("使用默认 MySQL 数据库: %s", dsn)
	}

	var err error
	//DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
	//	DisableForeignKeyConstraintWhenMigrating: true,
	//})
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.LogFatal("MySQL 连接失败: %v1", err)
	}
	if err != nil {
		log.LogFatal("数据库连接失败: %v1", err)
	}

	// 自动迁移实体表
	//err = DB.AutoMigrate(
	//	&po.TransactionRecord{},
	//	&po.Transaction{},
	//	&po.TransactionCreate{},
	//)
	//if err != nil {
	//	log.LogFatal("数据库表迁移失败: %v1", err)
	//}

	log.LogInfo("数据库初始化成功 (DSN: %s)", dsn)
}

// CloseDB 优雅关闭（main defer 调用，可选）
func CloseDB() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		sqlDB.Close()
	}
}
