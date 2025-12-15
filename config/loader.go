// config/loader.go
package config

import (
	"encoding/json"
	"go_fourmeme/entity/config"
	"io/ioutil"
	"os"

	"go_fourmeme/log"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

// LoadFromEnvAndFile 加载配置
// 优先级：env > .env > yaml/json 文件 > 默认值
func LoadFromEnvAndFile(configFile string) {
	// 加载 .env 文件
	_ = godotenv.Load()

	// 链配置从 env 覆盖
	if pk := os.Getenv("PRIVATE_KEY"); pk != "" {
		BSCChain.PrivateKey = pk
	}
	if url := os.Getenv("BSC_WS_URL"); url != "" {
		BSCChain.WSURL = url
	}
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		BSCChain.DBDSN = dsn
	}

	// 加载监听配置（yaml 或 json）
	if configFile == "" {
		log.LogInfo("无配置文件，使用默认配置")
		return
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.LogWarn("读取配置文件失败，使用默认: %v", err)
		return
	}

	var cfg struct {
		MonitorTargets []*config.MonitorTarget    `json:"monitor_targets" yaml:"monitor_targets"`
		SmartWallets   *config.SmartWalletsConfig `json:"smart_wallets" yaml:"smart_wallets"`
		Creators       *config.CreatorsConfig     `json:"creators" yaml:"creators"`
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		if jsonErr := json.Unmarshal(data, &cfg); jsonErr != nil {
			log.LogFatal("配置文件解析失败 (支持 yaml/json): yaml err: %v, json err: %v", err, jsonErr)
		}
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

	log.LogInfo("配置文件加载成功: %s", configFile)
}
