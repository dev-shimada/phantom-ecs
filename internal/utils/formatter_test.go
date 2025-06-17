package utils

import (
	"strings"
	"testing"

	"github.com/dev-shimada/phantom-ecs/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestFormatter_FormatJSON_ECSServices(t *testing.T) {
	formatter := NewFormatter()

	services := []models.ECSService{
		{
			ServiceName:    "web-service",
			ClusterName:    "test-cluster",
			Status:         "ACTIVE",
			TaskDefinition: "web-task:1",
			DesiredCount:   2,
			RunningCount:   2,
			LaunchType:     "FARGATE",
		},
		{
			ServiceName:    "api-service",
			ClusterName:    "test-cluster",
			Status:         "ACTIVE",
			TaskDefinition: "api-task:1",
			DesiredCount:   1,
			RunningCount:   1,
			LaunchType:     "EC2",
		},
	}

	result, err := formatter.FormatJSON(services)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "web-service")
	assert.Contains(t, result, "api-service")
	assert.Contains(t, result, "FARGATE")
	assert.Contains(t, result, "EC2")

	// JSONが正しい形式かチェック
	assert.True(t, strings.HasPrefix(result, "["))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(result), "]"))
}

func TestFormatter_FormatYAML_InspectionResult(t *testing.T) {
	formatter := NewFormatter()

	inspectionResult := models.InspectionResult{
		Service: models.ECSService{
			ServiceName:    "web-service",
			ClusterName:    "test-cluster",
			Status:         "ACTIVE",
			TaskDefinition: "web-task:1",
			DesiredCount:   2,
			RunningCount:   2,
			LaunchType:     "FARGATE",
		},
		TaskDefinition: models.ECSTaskDefinition{
			Family:      "web-task",
			Revision:    1,
			CPU:         "256",
			Memory:      "512",
			NetworkMode: "awsvpc",
			Status:      "ACTIVE",
		},
		NetworkConfig: &models.NetworkConfig{
			Subnets:        []string{"subnet-12345"},
			SecurityGroups: []string{"sg-abcdef"},
			AssignPublicIP: true,
		},
		Recommendations: []models.Recommendation{
			{
				Category:    "scaling",
				Title:       "Consider Auto Scaling",
				Description: "Enable ECS Service Auto Scaling",
				Priority:    "medium",
				Action:      "Configure Auto Scaling policies",
			},
		},
	}

	result, err := formatter.FormatYAML(inspectionResult)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "service:")
	assert.Contains(t, result, "task_definition:")
	assert.Contains(t, result, "network_config:")
	assert.Contains(t, result, "recommendations:")
	assert.Contains(t, result, "web-service")
	assert.Contains(t, result, "scaling")
}

func TestFormatter_FormatTable_ECSServices(t *testing.T) {
	formatter := NewFormatter()

	services := []models.ECSService{
		{
			ServiceName:    "web-service",
			ClusterName:    "test-cluster",
			Status:         "ACTIVE",
			TaskDefinition: "web-task:1",
			DesiredCount:   2,
			RunningCount:   2,
			LaunchType:     "FARGATE",
		},
		{
			ServiceName:    "api-service",
			ClusterName:    "prod-cluster",
			Status:         "ACTIVE",
			TaskDefinition: "api-task:1",
			DesiredCount:   1,
			RunningCount:   1,
			LaunchType:     "EC2",
		},
	}

	result, err := formatter.FormatTable(services)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// テーブルヘッダーのチェック
	assert.Contains(t, result, "SERVICE NAME")
	assert.Contains(t, result, "CLUSTER")
	assert.Contains(t, result, "STATUS")
	assert.Contains(t, result, "TASK DEFINITION")
	assert.Contains(t, result, "DESIRED")
	assert.Contains(t, result, "RUNNING")
	assert.Contains(t, result, "LAUNCH TYPE")

	// データ行のチェック
	assert.Contains(t, result, "web-service")
	assert.Contains(t, result, "api-service")
	assert.Contains(t, result, "test-cluster")
	assert.Contains(t, result, "prod-cluster")
	assert.Contains(t, result, "FARGATE")
	assert.Contains(t, result, "EC2")

	// テーブル形式のチェック（複数行であることを確認）
	lines := strings.Split(result, "\n")
	assert.True(t, len(lines) >= 3) // ヘッダー + 区切り線 + データ行以上
}

func TestFormatter_FormatTable_DeploymentResult(t *testing.T) {
	formatter := NewFormatter()

	deploymentResult := models.DeploymentResult{
		ServiceName:       "web-service-copy",
		ClusterName:       "target-cluster",
		TaskDefinitionArn: "arn:aws:ecs:us-west-2:123456789012:task-definition/web-task-copy:1",
		Success:           true,
		DryRun:            false,
	}

	result, err := formatter.FormatTable(deploymentResult)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "SERVICE NAME")
	assert.Contains(t, result, "CLUSTER")
	assert.Contains(t, result, "SUCCESS")
	assert.Contains(t, result, "DRY RUN")
	assert.Contains(t, result, "web-service-copy")
	assert.Contains(t, result, "target-cluster")
	assert.Contains(t, result, "true")
	assert.Contains(t, result, "false")
}

func TestFormatter_FormatCompact_ECSServices(t *testing.T) {
	formatter := NewFormatter()

	services := []models.ECSService{
		{
			ServiceName:  "web-service",
			ClusterName:  "test-cluster",
			Status:       "ACTIVE",
			DesiredCount: 2,
			RunningCount: 2,
		},
		{
			ServiceName:  "api-service",
			ClusterName:  "test-cluster",
			Status:       "ACTIVE",
			DesiredCount: 1,
			RunningCount: 0, // 不健全状態
		},
	}

	result, err := formatter.FormatCompact(services)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// コンパクト形式の基本チェック
	lines := strings.Split(result, "\n")
	assert.True(t, len(lines) >= 2)

	// サービス名とステータスが含まれることをチェック
	assert.Contains(t, result, "web-service")
	assert.Contains(t, result, "api-service")
	assert.Contains(t, result, "ACTIVE")
	assert.Contains(t, result, "2/2") // healthy
	assert.Contains(t, result, "0/1") // unhealthy
}

func TestFormatter_FormatWithOptions_JSON_Pretty(t *testing.T) {
	formatter := NewFormatter()

	service := models.ECSService{
		ServiceName: "web-service",
		ClusterName: "test-cluster",
		Status:      "ACTIVE",
	}

	options := FormatOptions{
		Format:       "json",
		PrettyPrint:  true,
		IncludeEmpty: false,
	}

	result, err := formatter.FormatWithOptions(service, options)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// プリティプリントされたJSONは複数行になる
	lines := strings.Split(result, "\n")
	assert.True(t, len(lines) > 1)

	// インデントが含まれることを確認
	assert.Contains(t, result, "  ") // インデント
}

func TestFormatter_FormatWithOptions_UnsupportedFormat(t *testing.T) {
	formatter := NewFormatter()

	service := models.ECSService{
		ServiceName: "web-service",
	}

	options := FormatOptions{
		Format: "xml", // サポートされていない形式
	}

	result, err := formatter.FormatWithOptions(service, options)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestFormatter_IsHealthyService(t *testing.T) {
	formatter := &Formatter{}

	healthyService := models.ECSService{
		Status:       "ACTIVE",
		DesiredCount: 2,
		RunningCount: 2,
	}

	unhealthyService := models.ECSService{
		Status:       "ACTIVE",
		DesiredCount: 2,
		RunningCount: 1,
	}

	inactiveService := models.ECSService{
		Status:       "INACTIVE",
		DesiredCount: 2,
		RunningCount: 2,
	}

	assert.True(t, formatter.IsHealthyService(healthyService))
	assert.False(t, formatter.IsHealthyService(unhealthyService))
	assert.False(t, formatter.IsHealthyService(inactiveService))
}
