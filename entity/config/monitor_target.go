// entity/config/monitor_target.go
package config

import "math/big"

// MonitorTarget 监听目标配置实体
type MonitorTarget struct {
	TokenName        string   `json:"token_name" yaml:"token_name"`
	TokenAddress     string   `json:"token_address" yaml:"token_address"`
	FourmemeManagers []string `json:"fourmeme_managers" yaml:"fourmeme_managers"`

	// 事件 Topic 完全隔离
	ERC20Topics    []string `json:"erc20_topics" yaml:"erc20_topics"`
	PancakeTopics  []string `json:"pancake_topics" yaml:"pancake_topics"`
	FourmemeTopics []string `json:"fourmeme_topics" yaml:"fourmeme_topics"`

	MethodIDsToMonitor []string `json:"method_ids_to_monitor" yaml:"method_ids_to_monitor"`

	// 交易策略
	BuyOnLiquidityAdd       bool       `json:"buy_on_liquidity_add" yaml:"buy_on_liquidity_add"`
	BuyAmountBNB            *big.Float `json:"buy_amount_bnb" yaml:"buy_amount_bnb"`
	SlippageTolerance       float64    `json:"slippage_tolerance" yaml:"slippage_tolerance"`
	TakeProfitMultiple      float64    `json:"take_profit_multiple" yaml:"take_profit_multiple"`
	StopLossMultiple        float64    `json:"stop_loss_multiple" yaml:"stop_loss_multiple"`
	TriggerOnSmartWalletBuy bool       `json:"trigger_on_smart_wallet_buy" yaml:"trigger_on_smart_wallet_buy"`
	TriggerOnCreatorAction  bool       `json:"trigger_on_creator_action" yaml:"trigger_on_creator_action"`
}
