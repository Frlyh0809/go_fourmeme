// config/creators.go
package config

type Creator struct {
	CreatorAddress     string   // 创建者地址（e.g., 部署合约的EOA）
	TopicsToMonitor    []string // 事件Topic（e.g., PairCreated, ContractDeployed如果可监听）
	MethodIDsToMonitor []string // Method ID（e.g., createPair）

	// 策略：检测到创建行为后触发什么
	OnCreateTokenAction  string // e.g., "monitor_new_token" - 开始监听新token
	OnAddLiquidityAction string // e.g., "buy" - 立即买入新token
}

type CreatorsConfig struct {
	Enabled  bool       // 是否启用整个creator监听
	Creators []*Creator // 创建者列表，如果为空则不监听
}

// 示例默认配置（可为空）
var DefaultCreators = &CreatorsConfig{
	Enabled: true,
	Creators: []*Creator{
		{
			CreatorAddress: "0xCreatorAddress1",
			TopicsToMonitor: []string{
				"0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c", // PairCreated
			},
			MethodIDsToMonitor:   []string{"0xYourCreateMethodID"},
			OnCreateTokenAction:  "monitor_new_token",
			OnAddLiquidityAction: "buy",
		},
		// 更多creator...
	},
}
