package config

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// /从配置文件中加载数据库相关的配置
func LoadFromEnvAndFile(configFile string) {
	godotenv.Load() // 加载.env

	// 加载链配置（从env覆盖默认）
	BSCChain.PrivateKey = os.Getenv("PRIVATE_KEY")
	// ... 其他覆盖

	// 动态加载监听配置（支持JSON或YAML）
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		// 使用默认配置
		return
	}

	var cfg struct {
		MonitorTargets []*MonitorTarget    `json:"monitor_targets" yaml:"monitor_targets"`
		SmartWallets   *SmartWalletsConfig `json:"smart_wallets" yaml:"smart_wallets"`
		Creators       *CreatorsConfig     `json:"creators" yaml:"creators"`
	}

	if err := json.Unmarshal(data, &cfg); err == nil {
		// JSON加载成功
	} else if err := yaml.Unmarshal(data, &cfg); err == nil {
		// YAML加载成功
	} else {
		panic("配置文件格式错误")
	}

	// 覆盖默认
	if cfg.MonitorTargets != nil {
		DefaultMonitorTargets = cfg.MonitorTargets
	}
	if cfg.SmartWallets != nil {
		DefaultSmartWallets = cfg.SmartWallets
	}
	if cfg.Creators != nil {
		DefaultCreators = cfg.Creators
	}
	// 覆盖DBDSN
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		//BSCChain.DBDSN = dsn
	}
}

// 获取私钥对象（推荐封装成局部函数，避免重复代码）
func GetPrivateKey() (*ecdsa.PrivateKey, error) {
	if BSCChain.PrivateKey == "" {
		return nil, fmt.Errorf("私钥未配置：PRIVATE_KEY 环境变量为空")
	}

	// 去除可能的 0x 前缀
	pkHex := strings.TrimPrefix(BSCChain.PrivateKey, "0x")

	pkBytes, err := hexutil.Decode("0x" + pkHex)
	if err != nil {
		return nil, fmt.Errorf("私钥解码失败: %v", err)
	}

	privateKey, err := crypto.ToECDSA(pkBytes)
	if err != nil {
		return nil, fmt.Errorf("私钥转换为 ECDSA 失败: %v", err)
	}

	return privateKey, nil
}
