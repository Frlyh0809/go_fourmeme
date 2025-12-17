// utils/seq.go 或直接放在 event 或 repository 包中
package utils // 根据需要调整包名

import "fmt"

// CalcTxUniqueSeq 生成 TxUniqueSeq 主键（block 9位 + txIndex 4位 + logIndex 4位）
func CalcTxUniqueSeq(blockNumber uint64, txIndex uint, logIndex uint) string {
	return fillZeroPadding(blockNumber, 9) +
		fillZeroPadding(uint64(txIndex), 4) +
		fillZeroPadding(uint64(logIndex), 4)
}

// fillZeroPadding 数字左补零到指定长度
func fillZeroPadding(num uint64, totalLength int) string {
	return fmt.Sprintf("%0*d", totalLength, num)
}
