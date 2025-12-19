package temp

import (
	"fmt"
	"go_fourmeme/client"
	"go_fourmeme/config"
	"go_fourmeme/database"
	"go_fourmeme/log"
	"go_fourmeme/manager"
	"go_fourmeme/utils"
	"math/big"
	"os"
	"testing"
	"time"
)

func before() {
	if err := os.Chdir(".."); err != nil { // 根据测试文件深度调整
		fmt.Printf("改变工作目录失败: %v1\n", err)
		os.Exit(1)
	}

	// 1. 加载配置
	configFile := "config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	config.LoadFromEnvAndFile(configFile)

	// 2. 初始化日志
	log.InitLogger()

	// 3. 加载 ABI 文件
	if err := utils.LoadABIs(); err != nil {
		log.LogFatal("ABI 加载失败: %v1", err)
	}

	// 4. 初始化数据库
	database.InitDB()
	client.InitBnbPriceCache()
	// 5. 连接客户端并设置全局
	ethClient, err := client.NewEthClientWithRetry(5, 5*time.Second)
	if err != nil {
		log.LogFatal("BSC 客户端连接失败: %v1", err)
	}
	manager.SetEthClient(ethClient)
	defer ethClient.Close()
}

func TestBuyToken(t *testing.T) {

}

func TestBuyTokenViaManager(t *testing.T) {
	before()
	monitorTargets := config.DefaultMonitorTargets[0]
	//res, err := BuyTokenViaManager(monitorTargets, "0xd22778601da716f3b774a0564e0cae0c3c484444")
	res1, err1 := BuyTokenViaManager2(monitorTargets, "0xd22778601da716f3b774a0564e0cae0c3c484444")
	log.LogInfo("res:%s err:%v1", res1, err1)

	res2, err2 := SellTokenViaManager2(monitorTargets, "0xd22778601da716f3b774a0564e0cae0c3c484444", utils.Mul10Pow(big.NewFloat(158817.9841695), big.NewInt(18)))

	log.LogInfo("res:%s err:%v1", res2, err2)
}
