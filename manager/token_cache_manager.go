// manager/token_cache_manager.go
package manager

import (
	"strings"
	"sync"
	"time"

	"go_fourmeme/database"
	"go_fourmeme/entity/po"
	"go_fourmeme/log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/patrickmn/go-cache"
)

// TokenCacheEntry 缓存中的 token 信息
type TokenCacheEntry struct {
	Creator       common.Address
	CreateTime    time.Time
	Level         string  // "normal", "suspected", "good" 等，可扩展
	BuyCount      int     // 买入笔数
	SellCount     int     // 卖出笔数
	BuyAmountBNB  float64 // 累计买入 BNB
	SellAmountBNB float64 // 累计卖出 BNB
}

var (
	tokenCache   *cache.Cache                // go-cache 实例
	blacklistSet = make(map[string]struct{}) // 黑名单集合
	whitelistSet = make(map[string]struct{}) // 白名单集合
	blMu         sync.RWMutex
	wlMu         sync.RWMutex
)

// InitTokenCache 初始化缓存 + 加载黑白名单到内存
func InitTokenCache() {
	// 10 分钟过期，30 分钟清理一次过期项
	tokenCache = cache.New(10*time.Minute, 30*time.Minute)

	// 加载黑名单
	var blacks []po.BlacklistCreator
	database.DB.Find(&blacks)
	for _, b := range blacks {
		blacklistSet[strings.ToLower(b.CreatorAddress)] = struct{}{}
	}

	// 加载白名单
	var whites []po.WhitelistCreator
	database.DB.Find(&whites)
	for _, w := range whites {
		whitelistSet[strings.ToLower(w.CreatorAddress)] = struct{}{}
	}

	log.LogInfo("Token 缓存初始化完成，黑名单 %d，白名单 %d", len(blacklistSet), len(whitelistSet))
}

// AddTokenToCache 新 token 创建时加入缓存
func AddTokenToCache(tokenAddr string, creator common.Address) {
	key := strings.ToLower(tokenAddr)
	tokenCache.Set(key, &TokenCacheEntry{
		Creator:      creator,
		CreateTime:   time.Now(),
		Level:        "normal",
		BuyCount:     0,
		SellCount:    0,
		BuyAmountBNB: 0,
	}, cache.DefaultExpiration)

	log.LogInfo("新 token 加入缓存: %s Creator: %s", tokenAddr[:10], creator.Hex())
}

// UpdateTradeStats 更新买卖统计（在发现买卖交易时调用）
func UpdateTradeStats(tokenAddr string, isBuy bool, amountBNB float64) {
	key := strings.ToLower(tokenAddr)
	val, found := tokenCache.Get(key)
	if !found {
		return
	}

	entry := val.(*TokenCacheEntry)
	if isBuy {
		entry.BuyCount++
		entry.BuyAmountBNB += amountBNB
	} else {
		entry.SellCount++
		entry.SellAmountBNB += amountBNB

		// 检查创建者是否在1分钟内卖出
		if strings.EqualFold(entry.Creator.Hex(), "seller_address") { // 传入 seller
			if time.Since(entry.CreateTime) <= time.Minute {
				AddToBlacklist(entry.Creator, tokenAddr)
				tokenCache.Delete(key) // 移除缓存
			}
		}
	}

	tokenCache.Set(key, entry, cache.DefaultExpiration)
}

// AddToBlacklist 添加黑名单
func AddToBlacklist(creator common.Address, tokenAddr string) {
	addr := strings.ToLower(creator.Hex())
	blMu.Lock()
	blacklistSet[addr] = struct{}{}
	blMu.Unlock()

	database.DB.Create(&po.BlacklistCreator{
		CreatorAddress: addr,
		Reason:         "creator sold within 1 minute after launch",
	})

	log.LogInfo("创建者加入黑名单: %s (Token: %s)", creator.Hex(), tokenAddr[:10])
}

// AddToWhitelist 添加白名单
func AddToWhitelist(creator common.Address, tokenAddr string) {
	addr := strings.ToLower(creator.Hex())
	wlMu.Lock()
	whitelistSet[addr] = struct{}{}
	wlMu.Unlock()

	database.DB.Create(&po.WhitelistCreator{
		CreatorAddress: addr,
		Reason:         "good distribution and market cap",
	})

	log.LogInfo("创建者加入白名单: %s (Token: %s)", creator.Hex(), tokenAddr[:10])
}

// IsBlacklisted / IsWhitelisted 检查
func IsBlacklisted(addr common.Address) bool {
	blMu.RLock()
	_, ok := blacklistSet[strings.ToLower(addr.Hex())]
	blMu.RUnlock()
	return ok
}

func IsWhitelisted(addr common.Address) bool {
	wlMu.RLock()
	_, ok := whitelistSet[strings.ToLower(addr.Hex())]
	wlMu.RUnlock()
	return ok
}
