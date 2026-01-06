package client

import (
	"encoding/json"
	"fmt"
	"go_fourmeme/config"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
)

// /https://docs.etherscan.io/resources/pro-endpoints
const bscScanBaseURL = "https://api.bscscan.com/api"

type APIResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

// Holder 持有人
type Holder struct {
	Address common.Address
	Balance string // string 存大数
}

// GetHistoricalTokenBalance 获取指定区块的 token 余额
func GetHistoricalTokenBalance(tokenAddr, holderAddr string, blockNo int64) (string, error) {
	params := map[string]string{
		"module":          "account",
		"action":          "tokenbalancehistory",
		"contractaddress": tokenAddr,
		"address":         holderAddr,
		"blockno":         strconv.FormatInt(blockNo, 10),
		"apikey":          config.BscScanAPIKey,
	}

	resp, err := http.Get(buildURL(params))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}

	if apiResp.Status != "1" {
		return "", fmt.Errorf("API error: %s", apiResp.Message)
	}

	return string(apiResp.Result), nil // 返回字符串余额（如 "1000000000000000000"）
}

// GetHistoricalTokenTotalSupply 获取指定区块的总供应量
func GetHistoricalTokenTotalSupply(tokenAddr string, blockNo int64) (string, error) {
	params := map[string]string{
		"module":          "stats",
		"action":          "tokensupplyhistory",
		"contractaddress": tokenAddr,
		"blockno":         strconv.FormatInt(blockNo, 10),
		"apikey":          config.BscScanAPIKey,
	}

	resp, err := http.Get(buildURL(params))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}

	if apiResp.Status != "1" {
		return "", fmt.Errorf("API error: %s", apiResp.Message)
	}

	return string(apiResp.Result), nil
}

// GetTopTokenHolders 获取 top 持有人（默认 top 100）
func GetTopTokenHolders(tokenAddr string, page, offset int) ([]Holder, int, error) {
	return getTokenHolders(tokenAddr, page, offset)
}

// GetTokenHolderList 获取持有人列表（分页）
func GetTokenHolderList(tokenAddr string, page, offset int) ([]Holder, int, error) {
	return getTokenHolders(tokenAddr, page, offset)
}

// GetTokenHolderCount 获取持有人数量
func GetTokenHolderCount(tokenAddr string) (int, error) {
	params := map[string]string{
		"module":          "token",
		"action":          "tokenholdercount",
		"contractaddress": tokenAddr,
		"apikey":          config.BscScanAPIKey,
	}

	resp, err := http.Get(buildURL(params))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return 0, err
	}

	if apiResp.Status != "1" {
		return 0, fmt.Errorf("API error: %s", apiResp.Message)
	}

	var count int
	if err := json.Unmarshal(apiResp.Result, &count); err != nil {
		return 0, err
	}

	return count, nil
}

// GetAddressTokenHolding 获取地址持有的 token 列表
func GetAddressTokenHolding(holderAddr string, page, offset int) ([]Holder, error) {
	params := map[string]string{
		"module":  "account",
		"action":  "addresstokenbalance",
		"address": holderAddr,
		"page":    strconv.Itoa(page),
		"offset":  strconv.Itoa(offset),
		"apikey":  config.BscScanAPIKey,
	}

	resp, err := http.Get(buildURL(params))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if apiResp.Status != "1" {
		return nil, fmt.Errorf("API error: %s", apiResp.Message)
	}

	var holders []Holder
	if err := json.Unmarshal(apiResp.Result, &holders); err != nil {
		return nil, err
	}

	return holders, nil
}

// 内部通用函数
func getTokenHolders(tokenAddr string, page, offset int) ([]Holder, int, error) {
	params := map[string]string{
		"module":          "token",
		"action":          "tokenholderlist",
		"contractaddress": tokenAddr,
		"page":            strconv.Itoa(page),
		"offset":          strconv.Itoa(offset),
		"apikey":          config.BscScanAPIKey,
	}

	resp, err := http.Get(buildURL(params))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, 0, err
	}

	if apiResp.Status != "1" {
		return nil, 0, fmt.Errorf("API error: %s", apiResp.Message)
	}

	var holders []struct {
		Address string `json:"TokenHolderAddress"`
		Balance string `json:"TokenHolderQuantity"`
	}
	if err := json.Unmarshal(apiResp.Result, &holders); err != nil {
		return nil, 0, err
	}

	var result []Holder
	for _, h := range holders {
		result = append(result, Holder{
			Address: common.HexToAddress(h.Address),
			Balance: h.Balance,
		})
	}

	total := len(holders) // 实际返回数量（可通过 count 接口精确）
	return result, total, nil
}

func buildURL(params map[string]string) string {
	url := bscScanBaseURL + "?"
	for k, v := range params {
		url += fmt.Sprintf("%s=%s&", k, v)
	}
	return url[:len(url)-1]
}
