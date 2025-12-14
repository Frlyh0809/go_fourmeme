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

	record.Timestamp = time.Now() // 统一设置时间戳

	if err := DB.Create(record).Error; err != nil {
		log.LogErrorMsg("保存交易记录失败: %v", err)
		return err
	}

	log.LogInfo("成功保存交易记录 [ID: %d | Tx: %s | Type: %s]", record.ID, record.TxHash[:10], record.Type)
	return nil
}

// GetTxRecordsByType 查询指定类型记录（示例，可扩展更多查询）
func GetTxRecordsByType(txType string) ([]po.TransactionRecord, error) {
	var records []po.TransactionRecord
	if err := DB.Where("type = ?", txType).Find(&records).Error; err != nil {
		log.LogErrorMsg("查询记录失败 [Type: %s]: %v", txType, err)
		return nil, err
	}

	return records, nil
}

// GetTxRecordByHash 通过 TxHash 查询单条记录
func GetTxRecordByHash(txHash string) (*po.TransactionRecord, error) {
	var record po.TransactionRecord
	if err := DB.Where("tx_hash = ?", txHash).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 未找到返回 nil
		}
		log.LogErrorMsg("查询记录失败 [Tx: %s]: %v", txHash[:10], err)
		return nil, err
	}

	return &record, nil
}

// UpdateTxRecordStatus 更新记录状态（e.g., pending -> success）
func UpdateTxRecordStatus(txHash, newStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status": newStatus,
	}
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}

	if err := DB.Model(&po.TransactionRecord{}).Where("tx_hash = ?", txHash).Updates(updates).Error; err != nil {
		log.LogErrorMsg("更新记录状态失败 [Tx: %s]: %v", txHash[:10], err)
		return err
	}

	log.LogInfo("更新记录状态成功 [Tx: %s | Status: %s]", txHash[:10], newStatus)
	return nil
}

// DeleteTxRecordByID 删除指定 ID 记录（可选，审计需求慎用）
func DeleteTxRecordByID(id uint) error {
	if err := DB.Delete(&po.TransactionRecord{}, id).Error; err != nil {
		log.LogErrorMsg("删除记录失败 [ID: %d]: %v", id, err)
		return err
	}

	log.LogInfo("删除记录成功 [ID: %d]", id)
	return nil
}
