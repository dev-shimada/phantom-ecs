package models

// InspectionResult はサービス調査結果を表す構造体
type InspectionResult struct {
	Service         ECSService        `json:"service" yaml:"service"`
	TaskDefinition  ECSTaskDefinition `json:"task_definition" yaml:"task_definition"`
	NetworkConfig   *NetworkConfig    `json:"network_config,omitempty" yaml:"network_config,omitempty"`
	Recommendations []Recommendation  `json:"recommendations" yaml:"recommendations"`
}

// NetworkConfig はネットワーク設定を表す構造体
type NetworkConfig struct {
	Subnets        []string `json:"subnets" yaml:"subnets"`
	SecurityGroups []string `json:"security_groups" yaml:"security_groups"`
	AssignPublicIP bool     `json:"assign_public_ip" yaml:"assign_public_ip"`
}

// Recommendation はレコメンデーション情報を表す構造体
type Recommendation struct {
	Category    string `json:"category" yaml:"category"`
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	Priority    string `json:"priority" yaml:"priority"` // high, medium, low
	Action      string `json:"action" yaml:"action"`
}
