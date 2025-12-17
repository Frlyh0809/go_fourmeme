package utils

import (
	"fmt"
	"math/big"
	"testing"
)

func TestDiv10Pow(t *testing.T) {
	//res := Div10Pow(big.NewInt(1), big.NewInt(10))
	res := DivFloat(big.NewFloat(0.009801980198), big.NewFloat(990524.0529))
	fmt.Println(BigFloatToString(res))
}
