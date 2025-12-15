// manager/position_manager.go
package manager

import (
	"strings"
	"sync"

	"go_fourmeme/entity"
	"go_fourmeme/log"
)

var (
	positionMu sync.RWMutex
	positions  = make(map[string]*entity.Position) // key: tokenAddr (小写标准化)
)

// AddPosition 添加或更新持仓（线程安全）
func AddPosition(pos *entity.Position) {
	if pos == nil {
		return
	}

	tokenKey := normalizeKey(pos.TokenAddr)

	positionMu.Lock()
	defer positionMu.Unlock()

	if existing, ok := positions[tokenKey]; ok {
		log.LogWarn("持仓已存在，更新: %s (旧Tx: %s → 新Tx: %s)",
			tokenKey[:10], existing.BuyTxHash[:10], pos.BuyTxHash[:10])
		// 可扩展为加权平均，这里简单覆盖（适合单次大额买入）
	}

	positions[tokenKey] = pos

	log.LogInfo("新增/更新持仓: %s | 投入 %.6f BNB | 获取 %s token | 止盈 %.1fx | 止损 %.1fx",
		tokenKey[:10],
		pos.BuyAmountBNB.Text('f', 6),
		pos.BuyTokenAmount.String(),
		pos.TargetProfitMult,
		pos.TargetLossMult)
}

// GetPosition 获取单个持仓
func GetPosition(tokenAddr string) (*entity.Position, bool) {
	positionMu.RLock()
	defer positionMu.RUnlock()

	pos, ok := positions[normalizeKey(tokenAddr)]
	return pos, ok
}

// GetAllPositions 获取所有持仓副本（用于监控）
func GetAllPositions() map[string]*entity.Position {
	positionMu.RLock()
	defer positionMu.RUnlock()

	copyMap := make(map[string]*entity.Position, len(positions))
	for k, v := range positions {
		copyMap[k] = v
	}
	return copyMap
}

// MarkAsSold 标记持仓已卖出
func MarkAsSold(tokenAddr string) {
	positionMu.Lock()
	defer positionMu.Unlock()

	key := normalizeKey(tokenAddr)
	if pos, ok := positions[key]; ok {
		pos.Sold = true
		log.LogInfo("持仓已标记为卖出: %s", key[:10])
	}
}

// DeletePosition 删除持仓（可选，清理用）
func DeletePosition(tokenAddr string) {
	positionMu.Lock()
	defer positionMu.Unlock()

	key := normalizeKey(tokenAddr)
	delete(positions, key)
	log.LogInfo("持仓已删除: %s", key[:10])
}

// normalizeKey 标准化地址 key（小写 + 去0x）
func normalizeKey(addr string) string {
	return strings.ToLower(strings.TrimPrefix(addr, "0x"))
}
