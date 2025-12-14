### 前述内容总结

基于之前的讨论，我们设计了一个使用GoLang开发的BSC链监听程序，专注于Fourmem token（或类似meme币）的交易监控和自动化操作。核心功能包括：
- **监听事件**：实时订阅BSC节点日志，捕获token创建、买卖、Transfer、进入二级市场（PancakeSwap）、创建池子、添加/移除流动性等行为。
- **交易逻辑**：支持一级（自定义池子）和二级市场（PancakeSwap）自动买卖，计算滑点（slippage），处理交易失败并日志记录。
- **配置解耦**：分化为链配置（ChainConfig：节点URL、ChainID、私钥等）、监听目标（MonitorTarget：token地址、事件Topic、Method ID、交易策略）、聪明钱包（SmartWalletsConfig：钱包地址、行为触发）、创建者（CreatorsConfig：creator地址、创建触发）。配置支持动态加载（env、JSON/YAML），每个部分可独立启用/禁用，便于灵活场景切换（如只监听token、不监听钱包）。
- **模块化架构**：main.go 作为入口，子模块包括config/（配置）、client/（BSC连接）、event/（监听&处理器）、trade/（买卖&滑点）、log/（日志）、utils/（工具）。强调并发安全、重试机制、性能优化。
- **扩展性**：支持多token/钱包监听，动态添加新目标，联动策略（e.g., 聪明钱包买入后跟随）。

现在，根据您的要求，添加**数据库模块**（database/），用于记录交易记录（e.g., 交易hash、类型、金额、时间、状态等）。这有助于审计、回溯和分析（如盈利统计）。数据库选择SQLite（本地文件，简单部署）或PostgreSQL（生产级，支持高并发）；使用GORM ORM简化CRUD操作。交易执行后立即记录，失败时也记录错误详情。整个架构保持解耦：如果不配置DB，则不启用记录功能。

### 重新设计的代码架构

这个架构是基于前述总结的完整版本，直接适用于实际开发。假设使用Go Modules（go mod init yourproject），依赖：
- `github.com/ethereum/go-ethereum`（BSC交互）
- `gorm.io/gorm` 和 `gorm.io/driver/sqlite`（或 `gorm.io/driver/postgres`）
- `github.com/sirupsen/logrus`（增强日志）
- `github.com/joho/godotenv`（env加载）
- `gopkg.in/yaml.v2` 和 `encoding/json`（配置加载）

#### 整体目录结构
```
yourproject/
├── main.go              # 程序入口，加载配置，启动监听
├── go.mod               # 依赖管理
├── go.sum
├── config.yaml          # 示例配置文件（可选）
├── .env                 # 私钥等敏感信息
├── database/            # 新增：数据库模块
│   ├── db.go            # DB连接和初始化
│   ├── models.go        # 数据模型（交易记录等）
│   └── repository.go    # CRUD操作
├── config/              # 配置模块（解耦分化）
│   ├── chain.go         # 链配置
│   ├── monitor.go       # Token监听配置
│   ├── smart_wallets.go # 聪明钱包配置
│   ├── creators.go      # Creator配置
│   ├── loader.go        # 配置加载（env + 文件）
│   └── types.go         # 公共类型
├── client/              # BSC客户端
│   └── client.go        # 连接和基本查询
├── event/               # 事件监听和处理器
│   ├── listener.go      # 订阅和监听（支持多类型）
│   └── handler.go       # 事件分类处理，触发交易/记录
├── trade/               # 交易逻辑
│   ├── buy_sell.go      # 买卖函数（一级/二级）
│   ├── slippage.go      # 滑点计算
│   └── liquidity.go     # 流动性操作
├── log/                 # 日志模块
│   └── logger.go        # 统一日志（info/error，文件输出）
└── utils/               # 工具函数
    ├── abi.go           # ABI加载和解析
    ├── error.go         # 错误重试机制
    └── math.go          # 大数转换（Wei/Ether）
```

#### 详细模块实现（代码片段）

1. **数据库模块（database/）** - 新增核心
    - **db.go**：连接和初始化。支持SQLite（默认，本地db文件）或PostgreSQL（配置DSN）。
      ```go
      // database/db.go
      package database
 
      import (
          "gorm.io/driver/sqlite" // 或 driver/postgres
          "gorm.io/gorm"
          "yourproject/config" // 导入配置
          "yourproject/log"
      )
 
      var DB *gorm.DB
 
      func InitDB(cfg *config.ChainConfig) {
          dsn := "transaction.db" // SQLite默认文件；PostgreSQL: cfg.DBDSN 如 "host=localhost user=gorm password=gorm dbname=gorm port=5432"
          var err error
          DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
          if err != nil {
              log.LogFatal("DB连接失败: %v", err)
          }
          // 自动迁移模型
          DB.AutoMigrate(&TransactionRecord{})
          log.LogInfo("DB初始化成功")
      }
      ```
    - **models.go**：交易记录模型（可扩展其他，如事件日志）。
      ```go
      // database/models.go
      package database
 
      import (
          "time"
          "math/big"
      )
 
      type TransactionRecord struct {
          ID        uint      `gorm:"primaryKey"`
          TxHash    string    `gorm:"uniqueIndex"` // 交易Hash
          Type      string    // e.g., "buy", "sell", "add_liquidity", "transfer"
          TokenAddr string    // 涉及的Token地址
          AmountIn  *big.Int  // 输入金额（Wei）
          AmountOut *big.Int  // 输出金额（Wei）
          Slippage  float64   // 实际滑点
          Status    string    // "success", "failed"
          ErrorMsg  string    // 失败原因
          Timestamp time.Time `gorm:"index"` // 时间戳
          // 扩展：WalletAddr（聪明钱包ID），CreatorAddr等
      }
      ```
    - **repository.go**：CRUD操作，交易后调用。
      ```go
      // database/repository.go
      package database
 
      func SaveTxRecord(record *TransactionRecord) error {
          record.Timestamp = time.Now()
          return DB.Create(record).Error
      }
 
      // 查询示例（可选，用于审计）
      func GetTxRecordsByType(txType string) ([]TransactionRecord, error) {
          var records []TransactionRecord
          return records, DB.Where("type = ?", txType).Find(&records).Error
      }
      ```

2. **配置模块（config/）** - 保持解耦，新增DB配置选项
    - **chain.go**：新增DBDSN字段（为空则用SQLite）。
      ```go
      // config/chain.go
      type ChainConfig struct {
          // ... 之前字段
          DBDSN string // PostgreSQL DSN，可为空
      }
      ```
    - **loader.go**：加载时检查DB配置。
      ```go
      // config/loader.go
      func LoadFromEnvAndFile(configFile string) {
          // ... 之前代码
          // 覆盖DBDSN
          if dsn := os.Getenv("DB_DSN"); dsn != "" {
              BSCChain.DBDSN = dsn
          }
      }
      ```
    - 其他（monitor/smart_wallets/creators）：不变，支持Enabled开关。

3. **客户端模块（client/）** - 不变
    - **client.go**：BSC连接，支持RPC和WS。

4. **事件模块（event/）** - 处理器中添加DB记录
    - **listener.go**：扩展支持多类型监听（token/钱包/creator）。
      ```go
      // event/listener.go
      func StartListener(client *ethclient.Client, addresses []common.Address, topics [][]common.Hash, handlerFunc func(types.Log)) {
          // ... 订阅FilterQuery
      }
 
      // 新增：聪明钱包监听
      func StartSmartWalletListener(client *ethclient.Client, cfg *config.SmartWalletsConfig) {
          if !cfg.Enabled { return }
          var addrs []common.Address
          for _, w := range cfg.Wallets { addrs = append(addrs, common.HexToAddress(w.WalletAddress)) }
          // 构建topics从cfg
          StartListener(client, addrs, topics, HandleEvent)
      }
 
      // 类似StartCreatorListener
      ```
    - **handler.go**：处理事件，触发交易，并记录DB。
      ```go
      // event/handler.go
      func HandleEvent(vLog types.Log) {
          // 解析事件
          switch vLog.Topics[0].Hex() {
          case "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef": // Transfer
              // 如果匹配聪明钱包/creator，触发trade.Buy/Sell
              txRecord := &po.TransactionRecord{Type: "transfer", TokenAddr: vLog.Address.Hex(), /* 填充数据 */}
              if err := trade.BuyToken(/* params */); err != nil {
                  txRecord.Status = "failed"
                  txRecord.ErrorMsg = err.Error()
                  log.LogFailure(err, "交易失败")
              } else {
                  txRecord.Status = "success"
              }
              database.SaveTxRecord(txRecord) // 记录到DB
          // 其他事件：PairCreated, Swap等
          }
      }
      ```

5. **交易模块（trade/）** - 执行后记录DB
    - **buy_sell.go**：买卖函数，成功/失败后调用DB。
      ```go
      // trade/buy_sell.go
      func BuyToken(client *ethclient.Client, amount *big.Int, tokenAddr string, slippage float64) (string, error) {
          // ... 构建tx，计算minOut = slippage.CalculateSlippage(...)
          // 发送tx
          if err := client.SendTransaction(...); err != nil {
              database.SaveTxRecord(&po.TransactionRecord{Type: "buy", TokenAddr: tokenAddr, Status: "failed", ErrorMsg: err.Error()})
              return "", err
          }
          txHash := signedTx.Hash().Hex()
          database.SaveTxRecord(&po.TransactionRecord{Type: "buy", TokenAddr: tokenAddr, AmountIn: amount, Slippage: slippage, Status: "success", TxHash: txHash})
          return txHash, nil
      }
      ```
    - **slippage.go** / **liquidity.go**：不变，集成到买卖中。

6. **日志模块（log/）** - 与DB互补
    - **logger.go**：使用logrus，支持文件输出。
      ```go
      // log/logger.go
      import "github.com/sirupsen/logrus"
 
      var Logger = logrus.New()
 
      func InitLogger() {
          Logger.SetOutput(os.Stdout) // 或文件
          Logger.SetLevel(logrus.InfoLevel)
      }
 
      func LogInfo(format string, args ...interface{}) {
          Logger.Infof(format, args...)
      }
 
      func LogFailure(err error, msg string) {
          Logger.Errorf("%s: %v", msg, err)
      }
      ```

7. **工具模块（utils/）** - 不变
    - ABI解析、重试（e.g., 指数退避重发tx）、大数处理。

#### 主程序（main.go）
```go
// main.go
package main

import (
    "yourproject/client"
    "yourproject/config"
    "yourproject/database"
    "yourproject/event"
    "yourproject/log"
)

func main() {
    config.LoadFromEnvAndFile("config.yaml")
    log.InitLogger()

    if config.BSCChain.DBDSN != "" || true { // 始终启用SQLite
        database.InitDB(config.BSCChain)
    }

    client, err := client.NewClient(config.BSCChain.WSURL)
    if err != nil {
        log.LogFatal("客户端失败: %v", err)
    }

    // 启动监听（根据配置灵活）
    for _, target := range config.DefaultMonitorTargets {
        go event.StartListener(client, []common.Address{common.HexToAddress(target.TokenAddress)}, /* topics */, event.HandleEvent)
    }
    if config.DefaultSmartWallets.Enabled {
        go event.StartSmartWalletListener(client, config.DefaultSmartWallets)
    }
    if config.DefaultCreators.Enabled {
        go event.StartCreatorListener(client, config.DefaultCreators)
    }

    select {} // 阻塞运行
}
```

#### 开发注意事项
- **部署**：运行`go run main.go`，确保.env有PRIVATE_KEY。测试用BSC Testnet（ChainID 97）。
- **错误处理**：所有tx发送添加重试（utils/error.go中实现，max 3次）。
- **性能**：监听高频事件时，用channel缓冲；DB用事务批量插入。
- **安全**：私钥加密存储，避免泄露。DB记录敏感数据时加密。
- **测试**：写单元测试（e.g., 测试滑点计算、DB插入）。集成测试用mock客户端。
- **扩展**：未来加Web UI查询DB记录，或导出CSV。

这个架构已足够完整，可直接复制到项目中开发。如果需要特定模块的完整代码、GORM迁移脚本或示例config.yaml，请提供更多细节！