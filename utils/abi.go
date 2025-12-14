// utils/abi.go
package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// 全局ABI变量
var (
	TokenManagerABI       abi.ABI // TokenManager1
	TokenManager2ABI      abi.ABI // 主Manager
	TokenManagerHelperABI abi.ABI
	ERC20ABI              abi.ABI
	PancakeRouterABI      abi.ABI
)

// LoadABIs 从根目录加载所有ABI文件
func LoadABIs() error {
	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}

	abiFiles := map[string]*abi.ABI{
		"resource/abi/TokenManager.lite.abi":   &TokenManagerABI,
		"resource/abi/TokenManager2.lite.abi":  &TokenManager2ABI,
		"resource/abi/TokenManagerHelper3.abi": &TokenManagerHelperABI,
		"resource/abi/ERC20.abi":               &ERC20ABI,
		"resource/abi/PancakeRouterV2.abi":     &PancakeRouterABI,
	}

	for filename, targetABI := range abiFiles {
		path := filepath.Join(rootDir, filename)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("[WARN] 加载ABI文件失败 %s: %v (如果不需要可忽略)", filename, err)
			continue
		}

		// 清理数据：去除BOM、换行等
		data = []byte(strings.TrimSpace(string(data)))

		if err := json.Unmarshal(data, targetABI); err != nil {
			log.Printf("[ERROR] 解析ABI JSON失败 %s: %v", filename, err)
			return err
		}
		log.Printf("[INFO] 成功加载ABI: %s", filename)
	}

	return nil
}

// GetABI 根据类型返回对应ABI
func GetABI(contractType string) *abi.ABI {
	switch contractType {
	case "TokenManager":
		return &TokenManagerABI
	case "TokenManager2":
		return &TokenManager2ABI
	case "TokenManagerHelper":
		return &TokenManagerHelperABI
	case "ERC20":
		return &ERC20ABI
	case "PancakeRouter":
		return &PancakeRouterABI
	default:
		return nil
	}
}
