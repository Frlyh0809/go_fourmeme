// database/repository.go
package database

import (
	"time"

	"go_fourmeme/entity/po"
	"go_fourmeme/log"

	"gorm.io/gorm"
)

// SaveTxRecord 保存交易记录
func SaveTxRecord(record *po.TransactionRecord) error {
	if record == nil {
		return nil
	}

	record.Timestamp = time.Now()

	if err := DB.Create(record).Error; err != nil {
		log.LogError("保存交易记录失败 (Tx: %s): %v1", record.TxHash[:10], err)
		return err
	}

	log.LogInfo("交易记录保存成功 [ID: %d | Tx: %s | Type: %s]", record.ID, record.TxHash[:10], record.Type)
	return nil
}

func SaveTransaction(tx *po.Transaction) error {
	result := DB.Create(tx)
	if result.Error != nil {
		log.LogError("数据库插入失败: %v1", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		log.LogWarn("插入 0 行，可能主键冲突 TxUniqueSeq: %s", tx.TxUniqueSeq)
	} else {
		log.LogInfo("插入成功 TxUniqueSeq: %s | Rows: %d", tx.TxUniqueSeq, result.RowsAffected)
	}
	return nil
}

func SaveTransactionCreate(tx *po.TransactionCreate) error {
	result := DB.Create(tx)
	if result.Error != nil {
		log.LogError("数据库插入失败 (create): %v1", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		log.LogWarn("插入 0 行 (create)，可能主键冲突: %s", tx.TxUniqueSeq)
	}
	return nil
}

// UpdateTxStatus 更新交易状态（例如 pending → success/failed）
func UpdateTxStatus(txHash, status, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}

	if err := DB.Model(&po.TransactionRecord{}).Where("tx_hash = ?", txHash).Updates(updates).Error; err != nil {
		log.LogError("更新交易状态失败 (Tx: %s): %v1", txHash[:10], err)
		return err
	}

	log.LogInfo("交易状态更新成功 [Tx: %s → %s]", txHash[:10], status)
	return nil
}

// GetTxRecordByHash 通过 TxHash 查询记录
func GetTxRecordByHash(txHash string) (*po.TransactionRecord, error) {
	var record po.TransactionRecord
	err := DB.Where("tx_hash = ?", txHash).First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// GetRecentRecords 获取最近 N 条记录（用于监控）
func GetRecentRecords(limit int) ([]po.TransactionRecord, error) {
	var records []po.TransactionRecord
	err := DB.Order("timestamp desc").Limit(limit).Find(&records).Error
	return records, err
}
