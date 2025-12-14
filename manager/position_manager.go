// manager/position_manager.go
package manager

import (
	//"math/big"
	"sync"
	//"time"

	"go_fourmeme/entity"
	"go_fourmeme/log"
)

var (
	positionMu sync.RWMutex
	positions  = make(map[string]*entity.Position) // key: tokenAddr (lowercase)
)

// AddPosition 添加或更新持仓（线程安全）
func AddPosition(pos *entity.Position) {
	if pos == nil {
		return
	}

	tokenAddr := normalizeAddr(pos.TokenAddr)

	positionMu.Lock()
	defer positionMu.Unlock()

	existing, exists := positions[tokenAddr]
	if exists {
		log.LogWarn("持仓已存在，更新记录: %s (原Tx: %s -> 新Tx: %s)", tokenAddr[:10], existing.BuyTxHash[:10], pos.BuyTxHash[:10])
		// 可选择合并持仓（加权平均价格），这里简单覆盖（适合单次买入策略）
	}

	positions[tokenAddr] = pos

	log.LogInfo("新增/更新持仓成功: %s | 投入 %.4f BNB | 获取 %s token | 止盈 %.1fx | 止损 %.1fx",
		tokenAddr[:10],
		pos.BuyAmountBNB.Text('f', 6),
		pos.BuyTokenAmount.String(),
		pos.TargetProfitMult,
		pos.TargetLossMult,
	)
}

// GetPosition 获取持仓（只读）
func GetPosition(tokenAddr string) (*entity.Position, bool) {
	positionMu.RLock()
	defer positionMu.RUnlock()

	pos, ok := positions[normalizeAddr(tokenAddr)]
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

// MarkAsSold 标记已卖出（盈亏监控调用）
func MarkAsSold(tokenAddr string) {
	positionMu.Lock()
	defer positionMu.Unlock()

	if pos, ok := positions[normalizeAddr(tokenAddr)]; ok {
		pos.Sold = true
		log.LogInfo("持仓标记为已卖出: %s", tokenAddr[:10])
	}
}

// normalizeAddr 统一地址格式（小写）
func normalizeAddr(addr string) string {
	if len(addr) > 2 && addr[:2] == "0x" {
		return addr
	}
	return addr
}
