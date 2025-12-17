// client/price_client.go
package client

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"go_fourmeme/log"
)

const (
	cryptoCompareURL = "https://min-api.cryptocompare.com/data/price?fsym=BNB&tsyms=USDT&api_key=7f58e0675d61e620b3ae94cc7d78c70d37528983b7136de41b305b76db32d7c9"
	updateInterval   = 10 * time.Minute // 缓存更新间隔
)

// 缓存结构
type bnbPriceCache struct {
	price     float64   // BNB in USDT
	timestamp time.Time // 最后更新时间
	mu        sync.RWMutex
}

var priceCache = &bnbPriceCache{}

// init 启动时立即查询一次，并开启后台定时更新
func InitBnbPriceCache() {
	go startPriceUpdater()
	// 启动时同步查询一次
	updateBNBPrice()
}

// GetBNBPriceUSDT 获取 BNB 对 USDT 价格（优先缓存）
func GetBNBPriceUSDT() float64 {
	priceCache.mu.RLock()
	if time.Since(priceCache.timestamp) < updateInterval && priceCache.price > 0 {
		defer priceCache.mu.RUnlock()
		return priceCache.price
	}
	priceCache.mu.RUnlock()

	// 缓存过期或无数据，实时查询
	if updateBNBPrice() {
		priceCache.mu.RLock()
		defer priceCache.mu.RUnlock()
		return priceCache.price
	}

	// 查询失败返回 0（调用方可处理）
	return 0
}

// updateBNBPrice 实时查询并更新缓存
func updateBNBPrice() bool {
	resp, err := http.Get(cryptoCompareURL)
	if err != nil {
		log.LogError("查询 BNB 价格失败: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.LogError("BNB 价格 API 返回非 200: %d", resp.StatusCode)
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.LogError("读取 BNB 价格响应失败: %v", err)
		return false
	}

	var result map[string]float64
	if err := json.Unmarshal(body, &result); err != nil {
		log.LogError("解析 BNB 价格 JSON 失败: %v | body: %s", err, string(body))
		return false
	}

	usdtPrice, ok := result["USDT"]
	if !ok || usdtPrice <= 0 {
		log.LogError("BNB 价格响应格式错误: %v", result)
		return false
	}

	// 更新缓存
	priceCache.mu.Lock()
	priceCache.price = usdtPrice
	priceCache.timestamp = time.Now()
	priceCache.mu.Unlock()

	log.LogInfo("BNB 价格更新成功: 1 BNB = %.4f USDT", usdtPrice)
	return true
}

// startPriceUpdater 后台定时更新价格
func startPriceUpdater() {
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			updateBNBPrice()
		}
	}
}
