package inspector_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/dev-shimada/phantom-ecs/internal/inspector"
	"github.com/dev-shimada/phantom-ecs/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockECSClient はECSクライアントのモック
type MockECSClient struct {
	mock.Mock
}

func (m *MockECSClient) ListClusters(ctx context.Context, input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*ecs.ListClustersOutput), args.Error(1)
}

func (m *MockECSClient) ListServices(ctx context.Context, input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*ecs.ListServicesOutput), args.Error(1)
}

func (m *MockECSClient) DescribeServices(ctx context.Context, input *ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*ecs.DescribeServicesOutput), args.Error(1)
}

func (m *MockECSClient) DescribeTaskDefinition(ctx context.Context, input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*ecs.DescribeTaskDefinitionOutput), args.Error(1)
}

func (m *MockECSClient) CreateService(ctx context.Context, input *ecs.CreateServiceInput) (*ecs.CreateServiceOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*ecs.CreateServiceOutput), args.Error(1)
}

func (m *MockECSClient) RegisterTaskDefinition(ctx context.Context, input *ecs.RegisterTaskDefinitionInput) (*ecs.RegisterTaskDefinitionOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*ecs.RegisterTaskDefinitionOutput), args.Error(1)
}

func TestInspector_InspectService_Success(t *testing.T) {
	mockClient := new(MockECSClient)
	inspector := inspector.NewInspector(mockClient)

	ctx := context.Background()
	serviceName := "web-service"
	clusterName := "test-cluster"

	// モックの設定 - サービス詳細取得
	mockClient.On("DescribeServices", ctx, &ecs.DescribeServicesInput{
		Cluster:  &clusterName,
		Services: []string{serviceName},
	}).Return(
		&ecs.DescribeServicesOutput{
			Services: []types.Service{
				{
					ServiceName:    stringPtr("web-service"),
					ServiceArn:     stringPtr("arn:aws:ecs:us-west-2:123456789012:service/test-cluster/web-service"),
					ClusterArn:     stringPtr("arn:aws:ecs:us-west-2:123456789012:cluster/test-cluster"),
					TaskDefinition: stringPtr("web-task:1"),
					DesiredCount:   2,
					RunningCount:   2,
					Status:         stringPtr("ACTIVE"),
					LaunchType:     types.LaunchTypeFargate,
					NetworkConfiguration: &types.NetworkConfiguration{
						AwsvpcConfiguration: &types.AwsVpcConfiguration{
							Subnets:        []string{"subnet-12345", "subnet-67890"},
							SecurityGroups: []string{"sg-abcdef"},
							AssignPublicIp: types.AssignPublicIpEnabled,
						},
					},
				},
			},
		}, nil)

	// モックの設定 - タスク定義詳細取得
	mockClient.On("DescribeTaskDefinition", ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: stringPtr("web-task:1"),
	}).Return(
		&ecs.DescribeTaskDefinitionOutput{
			TaskDefinition: &types.TaskDefinition{
				TaskDefinitionArn:       stringPtr("arn:aws:ecs:us-west-2:123456789012:task-definition/web-task:1"),
				Family:                  stringPtr("web-task"),
				Revision:                1,
				Status:                  types.TaskDefinitionStatusActive,
				Cpu:                     stringPtr("256"),
				Memory:                  stringPtr("512"),
				NetworkMode:             types.NetworkModeAwsvpc,
				RequiresCompatibilities: []types.Compatibility{types.CompatibilityFargate},
				ContainerDefinitions: []types.ContainerDefinition{
					{
						Name:  stringPtr("web-container"),
						Image: stringPtr("nginx:latest"),
						PortMappings: []types.PortMapping{
							{
								ContainerPort: int32Ptr(80),
								Protocol:      types.TransportProtocolTcp,
							},
						},
					},
				},
			},
		}, nil)

	// テスト実行
	result, err := inspector.InspectService(ctx, serviceName, clusterName)

	// アサーション
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// サービス情報の検証
	assert.Equal(t, "web-service", result.Service.ServiceName)
	assert.Equal(t, "test-cluster", result.Service.ClusterName)
	assert.Equal(t, "web-task:1", result.Service.TaskDefinition)
	assert.Equal(t, int32(2), result.Service.DesiredCount)
	assert.Equal(t, int32(2), result.Service.RunningCount)
	assert.Equal(t, "ACTIVE", result.Service.Status)
	assert.Equal(t, "FARGATE", result.Service.LaunchType)

	// タスク定義情報の検証
	assert.Equal(t, "web-task", result.TaskDefinition.Family)
	assert.Equal(t, 1, result.TaskDefinition.Revision)
	assert.Equal(t, "256", result.TaskDefinition.CPU)
	assert.Equal(t, "512", result.TaskDefinition.Memory)
	assert.Equal(t, "awsvpc", result.TaskDefinition.NetworkMode)

	// ネットワーク設定の検証
	assert.NotNil(t, result.NetworkConfig)
	assert.Len(t, result.NetworkConfig.Subnets, 2)
	assert.Contains(t, result.NetworkConfig.Subnets, "subnet-12345")
	assert.Contains(t, result.NetworkConfig.Subnets, "subnet-67890")
	assert.Len(t, result.NetworkConfig.SecurityGroups, 1)
	assert.Contains(t, result.NetworkConfig.SecurityGroups, "sg-abcdef")
	assert.True(t, result.NetworkConfig.AssignPublicIP)

	// レコメンデーションの検証
	assert.NotNil(t, result.Recommendations)
	assert.NotEmpty(t, result.Recommendations)

	mockClient.AssertExpectations(t)
}

func TestInspector_InspectService_ServiceNotFound(t *testing.T) {
	mockClient := new(MockECSClient)
	inspector := inspector.NewInspector(mockClient)

	ctx := context.Background()
	serviceName := "non-existent-service"
	clusterName := "test-cluster"

	// モックの設定 - サービスが見つからない
	mockClient.On("DescribeServices", ctx, &ecs.DescribeServicesInput{
		Cluster:  &clusterName,
		Services: []string{serviceName},
	}).Return(
		&ecs.DescribeServicesOutput{
			Services: []types.Service{}, // 空のサービス一覧
		}, nil)

	// テスト実行
	result, err := inspector.InspectService(ctx, serviceName, clusterName)

	// アサーション
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "service not found")

	mockClient.AssertExpectations(t)
}

func TestInspector_AnalyzeTaskDefinition_Success(t *testing.T) {
	mockClient := new(MockECSClient)
	inspector := inspector.NewInspector(mockClient)

	ctx := context.Background()
	taskDefArn := "web-task:1"

	// モックの設定
	mockClient.On("DescribeTaskDefinition", ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: &taskDefArn,
	}).Return(
		&ecs.DescribeTaskDefinitionOutput{
			TaskDefinition: &types.TaskDefinition{
				TaskDefinitionArn:       stringPtr("arn:aws:ecs:us-west-2:123456789012:task-definition/web-task:1"),
				Family:                  stringPtr("web-task"),
				Revision:                1,
				Status:                  types.TaskDefinitionStatusActive,
				Cpu:                     stringPtr("256"),
				Memory:                  stringPtr("512"),
				NetworkMode:             types.NetworkModeAwsvpc,
				RequiresCompatibilities: []types.Compatibility{types.CompatibilityFargate},
				ContainerDefinitions: []types.ContainerDefinition{
					{
						Name:  stringPtr("web-container"),
						Image: stringPtr("nginx:latest"),
						PortMappings: []types.PortMapping{
							{
								ContainerPort: int32Ptr(80),
								Protocol:      types.TransportProtocolTcp,
							},
						},
					},
				},
			},
		}, nil)

	// テスト実行
	result, err := inspector.AnalyzeTaskDefinition(ctx, taskDefArn)

	// アサーション
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "web-task", result.Family)
	assert.Equal(t, 1, result.Revision)
	assert.Equal(t, "256", result.CPU)
	assert.Equal(t, "512", result.Memory)
	assert.Equal(t, "awsvpc", result.NetworkMode)

	mockClient.AssertExpectations(t)
}

func TestInspector_GenerateRecommendations_HealthyService(t *testing.T) {
	inspector := &inspector.Inspector{}

	service := models.ECSService{
		ServiceName:  "web-service",
		Status:       "ACTIVE",
		DesiredCount: 2,
		RunningCount: 2,
		LaunchType:   "FARGATE",
	}

	taskDef := models.ECSTaskDefinition{
		CPU:    "256",
		Memory: "512",
	}

	recommendations := inspector.GenerateRecommendations(service, taskDef)

	// 健全なサービスでも基本的なレコメンデーションは提供される
	assert.NotEmpty(t, recommendations)

	// 具体的なレコメンデーション内容をチェック
	hasScalingRecommendation := false
	hasSecurityRecommendation := false

	for _, rec := range recommendations {
		if rec.Category == "scaling" {
			hasScalingRecommendation = true
		}
		if rec.Category == "security" {
			hasSecurityRecommendation = true
		}
	}

	assert.True(t, hasScalingRecommendation)
	assert.True(t, hasSecurityRecommendation)
}

func TestInspector_GenerateRecommendations_UnhealthyService(t *testing.T) {
	inspector := &inspector.Inspector{}

	service := models.ECSService{
		ServiceName:  "failing-service",
		Status:       "ACTIVE",
		DesiredCount: 3,
		RunningCount: 1, // 不健全状態
		LaunchType:   "EC2",
	}

	taskDef := models.ECSTaskDefinition{
		CPU:    "128", // 小さいCPU
		Memory: "256", // 小さいメモリ
	}

	recommendations := inspector.GenerateRecommendations(service, taskDef)

	assert.NotEmpty(t, recommendations)

	// 不健全なサービスには追加のレコメンデーションがある
	hasHealthRecommendation := false
	hasResourceRecommendation := false

	for _, rec := range recommendations {
		if rec.Category == "health" {
			hasHealthRecommendation = true
		}
		if rec.Category == "resources" {
			hasResourceRecommendation = true
		}
	}

	assert.True(t, hasHealthRecommendation)
	assert.True(t, hasResourceRecommendation)
}

// ヘルパー関数
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}
