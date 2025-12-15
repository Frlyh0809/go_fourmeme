// entity/config/creators.go
package config

// Creator 单个创建者配置
type Creator struct {
	CreatorAddress       string `json:"creator_address" yaml:"creator_address"`
	OnCreateTokenAction  string `json:"on_create_token_action" yaml:"on_create_token_action"`
	OnAddLiquidityAction string `json:"on_add_liquidity_action" yaml:"on_add_liquidity_action"`
}

// CreatorsConfig 创建者组配置
type CreatorsConfig struct {
	Enabled  bool       `json:"enabled" yaml:"enabled"`
	Creators []*Creator `json:"creators" yaml:"creators"`
}
