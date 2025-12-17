package utils

import (
	"math/big"
)

func BigIntToString(v *big.Int) string {
	if v == nil {
		return "0"
	}
	return v.String()
}
func BigFloatToString(v *big.Float) string {
	if v == nil {
		return "0"
	}
	return v.String()
}

// Div10Pow 安全地除以 10^pow（v != nil 时返回 v / 10^pow，否则返回 0）
func Div10Pow(v *big.Int, pow *big.Int) *big.Float {
	if v == nil {
		return big.NewFloat(0)
	}
	if pow == nil || pow.Sign() == 0 {
		return new(big.Float).SetInt(v) // pow=0，直接返回原值
	}

	// 构造 10^pow
	ten := big.NewInt(10)
	divisor := new(big.Int).Exp(ten, pow, nil)

	return new(big.Float).Quo(new(big.Float).SetInt(v), new(big.Float).SetInt(divisor))
}

// DivInt 安全地进行 a / b（a、b 都不为 nil 且 b != 0 时返回结果，否则返回 0）
func DivInt(a, b *big.Int) *big.Float {
	if a == nil || b == nil || b.Sign() == 0 {
		return big.NewFloat(0)
	}
	return new(big.Float).Quo(new(big.Float).SetInt(a), new(big.Float).SetInt(b))
}

func DivFloat(a, b *big.Float) *big.Float {
	if a == nil || b == nil || b.Sign() == 0 {
		return big.NewFloat(0)
	}
	return new(big.Float).Quo(a, b)
}
