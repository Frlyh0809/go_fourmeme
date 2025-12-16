package utils

import "math/big"

// splitDataToWords 将 log.Data 按 32 字节分块，返回 []*big.Int 数组
func SplitDataToWords(data []byte) []*big.Int {
	words := make([]*big.Int, 0, len(data)/32)
	for i := 0; i < len(data); i += 32 {
		end := i + 32
		if end > len(data) {
			end = len(data)
		}
		word := new(big.Int).SetBytes(data[i:end])
		words = append(words, word)
	}
	return words
}
