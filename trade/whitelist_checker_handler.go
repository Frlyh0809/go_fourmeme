package trade

import (
	"errors"
	"fmt"
	"math/big"
	_ "strconv"
	"time"

	"go_fourmeme/client"
	"go_fourmeme/log"

	"github.com/ethereum/go-ethereum/common"
)

const (
	MinHolders      = 20
	MaxTop5Percent  = 30 // top5 占比 ≤30%
	MinMarketCapUSD = 5000
)

// CheckWhitelistConditions 检查是否满足白名单条件（可定时或在买卖频繁时调用）
func CheckWhitelistConditions(tokenAddr string, creator common.Address, tokenInfo client.TokenInfo, createTime time.Time) (bool, error) {
	if time.Since(createTime) > 10*time.Minute {
		return false, nil // 超过10分钟不检查
	}

	if tokenInfo.Status != client.StatusTrading || tokenInfo.LiquidityAdded {
		return false, nil
	}

	// 获取 BNB 价格
	bnbPrice := client.GetBNBPriceUSDT()
	if bnbPrice <= 0 {
		//log.LogWarn("BNB 价格获取失败，跳过白名单检查")
		return false, errors.New("invalid bnbPrice")
	}

	// 计算市值 = Funds (BNB) * BNB价格
	marketCapUSD := new(big.Float).Mul(new(big.Float).SetInt(tokenInfo.Funds), big.NewFloat(bnbPrice))
	marketCapUSD.Quo(marketCapUSD, big.NewFloat(1e18)) // wei → BNB

	mcFloat, _ := marketCapUSD.Float64()
	if mcFloat < MinMarketCapUSD {
		log.LogInfo("[未满足]白单未满足-市值 tokenAddr:%s mcFloat:%d", tokenAddr, mcFloat)
		return false, nil // 市值的 <5000 USD
	}

	// 获取持有人数和 top5
	top5, totalHolders, err := client.GetTokenHolderList(tokenAddr, 1, 20)
	if err != nil {
		//log.LogError("获取持有人失败: %v", err)
		return false, fmt.Errorf("获取持有人失败 err: %v", err)
	}

	if totalHolders <= MinHolders {
		log.LogInfo("[未满足]白单未满足-持有人数 tokenAddr:%s totalHolders:%d", tokenAddr, totalHolders)
		return false, nil // 持有人 ≤20
	}

	// top5 总持有
	top5Total := big.NewInt(0)
	for _, h := range top5 {
		balance, _ := new(big.Int).SetString(h.Balance, 10)

		top5Total.Add(top5Total, balance)
	}

	// 当前解锁数量 ≈ Offers (已售)
	unlocked := tokenInfo.Offers
	if unlocked.Sign() == 0 {
		return false, nil
	}

	top5Percent := new(big.Float).Quo(new(big.Float).SetInt(top5Total), new(big.Float).SetInt(unlocked))
	pFloat, _ := top5Percent.Float64()
	if pFloat*100 > MaxTop5Percent {
		log.LogInfo("[未满足]白单未满足-top5持有超30% tokenAddr:%s pFloat:%d", tokenAddr, pFloat)
		return false, nil // top5 >30%
	}
	return true, nil
	// 满足所有条件 → 白名单
	//AddToWhitelist(creator, tokenAddr)
}
