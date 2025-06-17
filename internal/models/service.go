package models

import (
	"strconv"
	"strings"
	"time"
)

// ECSService ECSサービス情報を表す構造体
type ECSService struct {
	ServiceName    string    `json:"service_name" yaml:"service_name"`
	ClusterName    string    `json:"cluster_name" yaml:"cluster_name"`
	Status         string    `json:"status" yaml:"status"`
	TaskDefinition string    `json:"task_definition" yaml:"task_definition"`
	DesiredCount   int32     `json:"desired_count" yaml:"desired_count"`
	RunningCount   int32     `json:"running_count" yaml:"running_count"`
	CreatedAt      time.Time `json:"created_at" yaml:"created_at"`
	LaunchType     string    `json:"launch_type" yaml:"launch_type"`
}

// IsHealthy サービスが健全状態かどうかを判定
func (s *ECSService) IsHealthy() bool {
	return s.Status == "ACTIVE" && s.DesiredCount == s.RunningCount
}

// ECSTaskDefinition ECSタスク定義情報を表す構造体
type ECSTaskDefinition struct {
	TaskDefinitionArn  string   `json:"task_definition_arn" yaml:"task_definition_arn"`
	Family             string   `json:"family" yaml:"family"`
	Revision           int      `json:"revision" yaml:"revision"`
	Status             string   `json:"status" yaml:"status"`
	CPU                string   `json:"cpu" yaml:"cpu"`
	Memory             string   `json:"memory" yaml:"memory"`
	NetworkMode        string   `json:"network_mode" yaml:"network_mode"`
	RequiresAttributes []string `json:"requires_attributes" yaml:"requires_attributes"`
}

// GetFamilyAndRevision ARNからファミリー名とリビジョン番号を抽出
func (td *ECSTaskDefinition) GetFamilyAndRevision() (string, int) {
	if td.TaskDefinitionArn == "" {
		return "", 0
	}

	// ARN形式: arn:aws:ecs:region:account:task-definition/family:revision
	parts := strings.Split(td.TaskDefinitionArn, "/")
	if len(parts) < 2 {
		return "", 0
	}

	familyRevision := parts[len(parts)-1]
	familyParts := strings.Split(familyRevision, ":")
	if len(familyParts) < 2 {
		return familyRevision, 0
	}

	family := familyParts[0]
	revision, err := strconv.Atoi(familyParts[1])
	if err != nil {
		return family, 0
	}

	return family, revision
}

// ECSCluster ECSクラスター情報を表す構造体
type ECSCluster struct {
	ClusterName                       string `json:"cluster_name" yaml:"cluster_name"`
	ClusterArn                        string `json:"cluster_arn" yaml:"cluster_arn"`
	Status                            string `json:"status" yaml:"status"`
	RunningTasksCount                 int32  `json:"running_tasks_count" yaml:"running_tasks_count"`
	ActiveServicesCount               int32  `json:"active_services_count" yaml:"active_services_count"`
	RegisteredContainerInstancesCount int32  `json:"registered_container_instances_count" yaml:"registered_container_instances_count"`
}
