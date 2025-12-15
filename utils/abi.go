// utils/abi.go
package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"go_fourmeme/log"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	TokenManagerABI       abi.ABI // TokenManager1
	TokenManager2ABI      abi.ABI // 主 Manager
	TokenManagerHelperABI abi.ABI
	ERC20ABI              abi.ABI
	PancakeRouterABI      abi.ABI
)

// LoadABIs 从项目根目录加载所有 ABI 文件
func LoadABIs() error {
	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}

	abiMap := map[string]*abi.ABI{
		"resource/abi/TokenManager.lite.abi":   &TokenManagerABI,
		"resource/abi/TokenManager2.lite.abi":  &TokenManager2ABI,
		"resource/abi/TokenManagerHelper3.abi": &TokenManagerHelperABI,
		"resource/abi/ERC20.abi":               &ERC20ABI,
		"resource/abi/PancakeRouterV2.abi":     &PancakeRouterABI,
	}

	for filename, target := range abiMap {
		path := filepath.Join(rootDir, filename)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.LogWarn("ABI 文件未找到或读取失败: %s (%v)，部分功能可能不可用", filename, err)
			continue
		}

		// 清理数据
		cleanData := strings.TrimSpace(string(data))

		if err := json.Unmarshal([]byte(cleanData), target); err != nil {
			log.LogError("解析 ABI JSON 失败: %s (%v)", filename, err)
			return err
		}

		log.LogInfo("成功加载 ABI: %s", filename)
	}

	return nil
}

// GetABI 根据类型返回对应 ABI
func GetABI(abiType string) *abi.ABI {
	switch abiType {
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

//abiFiles := map[string]*abi.ABI{
//	"resource/abi/TokenManager.lite.abi":   &TokenManagerABI,
//	"resource/abi/TokenManager2.lite.abi":  &TokenManager2ABI,
//	"resource/abi/TokenManagerHelper3.abi": &TokenManagerHelperABI,
//	"resource/abi/ERC20.abi":               &ERC20ABI,
//	"resource/abi/PancakeRouterV2.abi":     &PancakeRouterABI,
//}
