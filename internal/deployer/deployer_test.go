package deployer_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/dev-shimada/phantom-ecs/internal/deployer"
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

func TestDeployer_DeployService_Success(t *testing.T) {
	mockClient := new(MockECSClient)
	deployer := deployer.NewDeployer(mockClient)

	ctx := context.Background()

	// テスト用のInspectionResult
	inspectionResult := &models.InspectionResult{
		Service: models.ECSService{
			ServiceName:    "web-service",
			ClusterName:    "source-cluster",
			TaskDefinition: "web-task:1",
			DesiredCount:   2,
			LaunchType:     "FARGATE",
			Status:         "ACTIVE",
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
	}

	targetCluster := "target-cluster"
	newServiceName := "web-service-copy"

	// モックの設定 - タスク定義登録
	mockClient.On("RegisterTaskDefinition", ctx, mock.MatchedBy(func(input *ecs.RegisterTaskDefinitionInput) bool {
		return *input.Family == "web-task-copy"
	})).Return(
		&ecs.RegisterTaskDefinitionOutput{
			TaskDefinition: &types.TaskDefinition{
				TaskDefinitionArn: func() *string { s := "arn:aws:ecs:us-west-2:123456789012:task-definition/web-task-copy:1"; return &s }(),
				Family:            func() *string { s := "web-task-copy"; return &s }(),
				Revision:          1,
			},
		}, nil)

	// モックの設定 - サービス作成
	mockClient.On("CreateService", ctx, mock.MatchedBy(func(input *ecs.CreateServiceInput) bool {
		return *input.ServiceName == newServiceName && *input.Cluster == targetCluster
	})).Return(
		&ecs.CreateServiceOutput{
			Service: &types.Service{
				ServiceName: &newServiceName,
				ServiceArn: func() *string {
					s := "arn:aws:ecs:us-west-2:123456789012:service/target-cluster/web-service-copy"
					return &s
				}(),
				ClusterArn: func() *string { s := "arn:aws:ecs:us-west-2:123456789012:cluster/target-cluster"; return &s }(),
			},
		}, nil)

	// テスト実行
	result, err := deployer.DeployService(ctx, inspectionResult, targetCluster, newServiceName, false)

	// アサーション
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newServiceName, result.ServiceName)
	assert.Equal(t, targetCluster, result.ClusterName)
	assert.Equal(t, "arn:aws:ecs:us-west-2:123456789012:task-definition/web-task-copy:1", result.TaskDefinitionArn)
	assert.True(t, result.Success)

	mockClient.AssertExpectations(t)
}

func TestDeployer_DeployService_DryRun(t *testing.T) {
	mockClient := new(MockECSClient)
	deployer := deployer.NewDeployer(mockClient)

	ctx := context.Background()

	inspectionResult := &models.InspectionResult{
		Service: models.ECSService{
			ServiceName:    "web-service",
			ClusterName:    "source-cluster",
			TaskDefinition: "web-task:1",
			DesiredCount:   2,
			LaunchType:     "FARGATE",
			Status:         "ACTIVE",
		},
		TaskDefinition: models.ECSTaskDefinition{
			Family:      "web-task",
			Revision:    1,
			CPU:         "256",
			Memory:      "512",
			NetworkMode: "awsvpc",
			Status:      "ACTIVE",
		},
	}

	targetCluster := "target-cluster"
	newServiceName := "web-service-copy"

	// Dry runの場合はAWS APIは呼ばれない

	// テスト実行
	result, err := deployer.DeployService(ctx, inspectionResult, targetCluster, newServiceName, true)

	// アサーション
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newServiceName, result.ServiceName)
	assert.Equal(t, targetCluster, result.ClusterName)
	assert.True(t, result.DryRun)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Operations)

	// AWS APIが呼ばれていないことを確認
	mockClient.AssertNotCalled(t, "RegisterTaskDefinition")
	mockClient.AssertNotCalled(t, "CreateService")
}

func TestDeployer_CloneTaskDefinition_Success(t *testing.T) {
	mockClient := new(MockECSClient)
	deployer := deployer.NewDeployer(mockClient)

	ctx := context.Background()

	sourceTaskDef := models.ECSTaskDefinition{
		Family:      "web-task",
		Revision:    1,
		CPU:         "256",
		Memory:      "512",
		NetworkMode: "awsvpc",
		Status:      "ACTIVE",
	}

	newFamily := "web-task-copy"

	// モックの設定
	mockClient.On("RegisterTaskDefinition", ctx, mock.MatchedBy(func(input *ecs.RegisterTaskDefinitionInput) bool {
		return *input.Family == newFamily &&
			*input.Cpu == "256" &&
			*input.Memory == "512"
	})).Return(
		&ecs.RegisterTaskDefinitionOutput{
			TaskDefinition: &types.TaskDefinition{
				TaskDefinitionArn: func() *string { s := "arn:aws:ecs:us-west-2:123456789012:task-definition/web-task-copy:1"; return &s }(),
				Family:            func() *string { return &newFamily }(),
				Revision:          1,
			},
		}, nil)

	// テスト実行
	result, err := deployer.CloneTaskDefinition(ctx, sourceTaskDef, newFamily)

	// アサーション
	assert.NoError(t, err)
	assert.Equal(t, "arn:aws:ecs:us-west-2:123456789012:task-definition/web-task-copy:1", result)

	mockClient.AssertExpectations(t)
}

func TestDeployer_CustomizeService_BasicCustomization(t *testing.T) {
	deployer := &deployer.Deployer{}

	sourceService := models.ECSService{
		ServiceName:    "web-service",
		ClusterName:    "source-cluster",
		TaskDefinition: "web-task:1",
		DesiredCount:   2,
		LaunchType:     "FARGATE",
		Status:         "ACTIVE",
	}

	customization := models.DeploymentCustomization{
		NewServiceName: "web-service-copy",
		TargetCluster:  "target-cluster",
		DesiredCount:   &[]int32{3}[0],
		LaunchType:     "EC2",
	}

	result := deployer.CustomizeService(sourceService, customization)

	assert.Equal(t, "web-service-copy", result.ServiceName)
	assert.Equal(t, "target-cluster", result.ClusterName)
	assert.Equal(t, int32(3), result.DesiredCount)
	assert.Equal(t, "EC2", result.LaunchType)
	// 元のタスク定義は保持される
	assert.Equal(t, "web-task:1", result.TaskDefinition)
}

func TestDeployer_ValidateDeployment_Success(t *testing.T) {
	deployer := &deployer.Deployer{}

	inspectionResult := &models.InspectionResult{
		Service: models.ECSService{
			ServiceName: "web-service",
			Status:      "ACTIVE",
		},
		TaskDefinition: models.ECSTaskDefinition{
			Family: "web-task",
			Status: "ACTIVE",
		},
	}

	targetCluster := "target-cluster"
	newServiceName := "web-service-copy"

	err := deployer.ValidateDeployment(inspectionResult, targetCluster, newServiceName)

	assert.NoError(t, err)
}

func TestDeployer_ValidateDeployment_InvalidSource(t *testing.T) {
	deployer := &deployer.Deployer{}

	inspectionResult := &models.InspectionResult{
		Service: models.ECSService{
			ServiceName: "web-service",
			Status:      "INACTIVE", // 無効なステータス
		},
		TaskDefinition: models.ECSTaskDefinition{
			Family: "web-task",
			Status: "ACTIVE",
		},
	}

	targetCluster := "target-cluster"
	newServiceName := "web-service-copy"

	err := deployer.ValidateDeployment(inspectionResult, targetCluster, newServiceName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source service is not active")
}

func TestDeployer_ValidateDeployment_EmptyTargetCluster(t *testing.T) {
	deployer := &deployer.Deployer{}

	inspectionResult := &models.InspectionResult{
		Service: models.ECSService{
			ServiceName: "web-service",
			Status:      "ACTIVE",
		},
		TaskDefinition: models.ECSTaskDefinition{
			Family: "web-task",
			Status: "ACTIVE",
		},
	}

	targetCluster := ""
	newServiceName := "web-service-copy"

	err := deployer.ValidateDeployment(inspectionResult, targetCluster, newServiceName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target cluster name cannot be empty")
}
