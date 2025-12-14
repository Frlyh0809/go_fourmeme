package trade

import "math/big"

func CalculateSlippage(input, reserveIn, reserveOut *big.Int, slippage float64) (*big.Int, error) {
	expected := new(big.Int).Mul(input, reserveOut)
	expected.Div(expected, new(big.Int).Add(reserveIn, input))
	minOut := new(big.Int).Mul(expected, big.NewInt(int64(100-slippage*100)))
	minOut.Div(minOut, big.NewInt(100))
	return minOut, nil
}
