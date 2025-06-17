package models

// DeploymentResult はデプロイメント結果を表す構造体
type DeploymentResult struct {
	ServiceName       string   `json:"service_name" yaml:"service_name"`
	ClusterName       string   `json:"cluster_name" yaml:"cluster_name"`
	TaskDefinitionArn string   `json:"task_definition_arn" yaml:"task_definition_arn"`
	Success           bool     `json:"success" yaml:"success"`
	DryRun            bool     `json:"dry_run" yaml:"dry_run"`
	Operations        []string `json:"operations,omitempty" yaml:"operations,omitempty"`
	Error             string   `json:"error,omitempty" yaml:"error,omitempty"`
}

// DeploymentCustomization はデプロイメントのカスタマイズオプションを表す構造体
type DeploymentCustomization struct {
	NewServiceName string  `json:"new_service_name" yaml:"new_service_name"`
	TargetCluster  string  `json:"target_cluster" yaml:"target_cluster"`
	DesiredCount   *int32  `json:"desired_count,omitempty" yaml:"desired_count,omitempty"`
	LaunchType     string  `json:"launch_type,omitempty" yaml:"launch_type,omitempty"`
	CPU            *string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory         *string `json:"memory,omitempty" yaml:"memory,omitempty"`
}
