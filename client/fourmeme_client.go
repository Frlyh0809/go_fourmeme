// client/fourmeme_client.go (追加或替换)
package client

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"go_fourmeme/config"
	"go_fourmeme/log"
	"go_fourmeme/manager"
	"go_fourmeme/utils"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// TokenStatus 枚举
type TokenStatus string

const (
	StatusUnknown         TokenStatus = "UNKNOWN"
	StatusTrading         TokenStatus = "TRADING"
	StatusAddingLiquidity TokenStatus = "ADDING_LIQUIDITY"
	StatusHalt            TokenStatus = "HALT"
	StatusCompleted       TokenStatus = "COMPLETED"
)

// TokenInfo 从 getTokenInfo 返回的详细信息
type TokenInfo struct {
	Version        *big.Int
	TokenManager   common.Address
	Quote          common.Address
	LastPrice      *big.Int
	TradingFeeRate *big.Int
	MinTradingFee  *big.Int
	LaunchTime     *big.Int
	Offers         *big.Int
	MaxOffers      *big.Int
	Funds          *big.Int
	MaxFunds       *big.Int
	LiquidityAdded bool
	Status         TokenStatus // 推导出的状态（非直接返回）
}

// GetTokenStatus 查询 token 状态（使用 getTokenInfo）
func GetTokenStatus(tokenAddr string) TokenInfo {
	ethClient := manager.GetEthClient()
	if ethClient == nil {
		log.LogError("ethclient 未初始化")
		return TokenInfo{Status: StatusUnknown}
	}

	helperABI := utils.GetABI("TokenManagerHelper")
	if helperABI == nil {
		log.LogError("TokenManagerHelper ABI 未加载")
		return TokenInfo{Status: StatusUnknown}
	}

	helperAddr := common.HexToAddress(config.TokenManagerHelper3)

	// Pack 输入
	input, err := helperABI.Pack("getTokenInfo", common.HexToAddress(tokenAddr))
	if err != nil {
		log.LogError("Pack getTokenInfo 失败: %v1", err)
		return TokenInfo{Status: StatusUnknown}
	}

	// Call
	output, err := ethClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &helperAddr,
		Data: input,
	}, nil)
	if err != nil {
		log.LogWarn("调用 getTokenInfo 失败 (Token: %s): %v1", tokenAddr[:10], err)
		return TokenInfo{Status: StatusUnknown}
	}

	// 定义结构体匹配 ABI 输出顺序
	var info struct {
		Version        *big.Int
		TokenManager   common.Address
		Quote          common.Address
		LastPrice      *big.Int
		TradingFeeRate *big.Int
		MinTradingFee  *big.Int
		LaunchTime     *big.Int
		Offers         *big.Int
		MaxOffers      *big.Int
		Funds          *big.Int
		MaxFunds       *big.Int
		LiquidityAdded bool
	}

	err = helperABI.UnpackIntoInterface(&info, "getTokenInfo", output)
	if err != nil {
		log.LogError("解包 getTokenInfo 失败: %v1", err)
		return TokenInfo{Status: StatusUnknown}
	}

	// 推导状态（Fourmeme 逻辑：LiquidityAdded == true 表示已毕业）
	status := StatusTrading
	if info.LiquidityAdded {
		status = StatusCompleted
	}
	// 其他状态（如 HALT）需额外查询或从 Funds/Offers 推断，这里简化

	//log.LogInfo("Token %s 状态查询成功: LiquidityAdded=%v1 | Offers=%s | Funds=%s",
	//	tokenAddr[:10], info.LiquidityAdded, info.Offers.String(), info.Funds.String())
	log.LogInfo(`Token %s 状态详情:
    Version:        %s
    TokenManager:   %s
    Quote:          %s
    LastPrice:      %s
    TradingFeeRate: %s
    MinTradingFee:  %s
    LaunchTime:     %s (%s)
    Offers:         %s
    MaxOffers:      %s
    Funds:          %s
    MaxFunds:       %s
    LiquidityAdded: %v1
    Status:         %s`,
		tokenAddr[:10],
		info.Version.String(),
		info.TokenManager.Hex(),
		info.Quote.Hex(),
		info.LastPrice.String(),
		info.TradingFeeRate.String(),
		info.MinTradingFee.String(),
		info.LaunchTime.String(),
		time.Unix(info.LaunchTime.Int64(), 0).Format("2006-01-02 15:04:05"),
		info.Offers.String(),
		info.MaxOffers.String(),
		info.Funds.String(),
		info.MaxFunds.String(),
		info.LiquidityAdded,
		status,
	)

	return TokenInfo{
		Version:        info.Version,
		TokenManager:   info.TokenManager,
		Quote:          info.Quote,
		LastPrice:      info.LastPrice,
		TradingFeeRate: info.TradingFeeRate,
		MinTradingFee:  info.MinTradingFee,
		LaunchTime:     info.LaunchTime,
		Offers:         info.Offers,
		MaxOffers:      info.MaxOffers,
		Funds:          info.Funds,
		MaxFunds:       info.MaxFunds,
		LiquidityAdded: info.LiquidityAdded,
		Status:         status,
	}
}

// LocalCalcMinAmountOut 本地计算 minAmountOut（绕过 calcTokenOut）
func LocalCalcMinAmountOut(info TokenInfo, amountInWei *big.Int, slippage float64) (*big.Int, error) {
	if amountInWei.Sign() == 0 {
		return big.NewInt(0), nil
	}

	if info.Offers.Sign() == 0 || info.Funds.Sign() == 0 {
		return nil, fmt.Errorf("Offers or Funds 为 0，无法本地计算")
	}

	// 当前价格 = Offers / Funds (token per BNB wei)
	currentPrice := new(big.Float).Quo(
		new(big.Float).SetInt(info.Offers),
		new(big.Float).SetInt(info.Funds),
	)

	// 预期 tokenOut ≈ amountInWei / currentPrice
	expectedOutFloat := new(big.Float).Quo(
		new(big.Float).SetInt(amountInWei),
		currentPrice,
	)

	// 转为 big.Int
	expectedOut := new(big.Int)
	expectedOutFloat.Int(expectedOut)

	// 滑点保护
	slippageFactor := big.NewFloat(1 - slippage)
	minOutFloat := new(big.Float).Mul(expectedOutFloat, slippageFactor)
	minOut := new(big.Int)
	minOutFloat.Int(minOut)

	log.LogInfo("本地滑点计算: expectedOut ≈ %s | minOut = %s (slippage %.2f%%)",
		utils.BigIntToString(expectedOut), utils.BigIntToString(minOut), slippage*100)

	return minOut, nil
}
