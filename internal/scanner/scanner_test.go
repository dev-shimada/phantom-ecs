package scanner_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/dev-shimada/phantom-ecs/internal/scanner"
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

func TestScanner_ScanServices_SingleCluster(t *testing.T) {
	mockClient := new(MockECSClient)
	scanner := scanner.NewScanner(mockClient)

	ctx := context.Background()
	clusterName := "test-cluster"

	// モックの設定 - サービス一覧取得
	mockClient.On("ListServices", ctx, &ecs.ListServicesInput{
		Cluster: &clusterName,
	}).Return(
		&ecs.ListServicesOutput{
			ServiceArns: []string{
				"arn:aws:ecs:us-west-2:123456789012:service/test-cluster/web-service",
				"arn:aws:ecs:us-west-2:123456789012:service/test-cluster/api-service",
			},
		}, nil)

	// モックの設定 - サービス詳細取得
	mockClient.On("DescribeServices", ctx, &ecs.DescribeServicesInput{
		Cluster: &clusterName,
		Services: []string{
			"arn:aws:ecs:us-west-2:123456789012:service/test-cluster/web-service",
			"arn:aws:ecs:us-west-2:123456789012:service/test-cluster/api-service",
		},
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
				},
				{
					ServiceName:    stringPtr("api-service"),
					ServiceArn:     stringPtr("arn:aws:ecs:us-west-2:123456789012:service/test-cluster/api-service"),
					ClusterArn:     stringPtr("arn:aws:ecs:us-west-2:123456789012:cluster/test-cluster"),
					TaskDefinition: stringPtr("api-task:1"),
					DesiredCount:   1,
					RunningCount:   1,
					Status:         stringPtr("ACTIVE"),
				},
			},
		}, nil)

	// テスト実行
	result, err := scanner.ScanServices(ctx, []string{clusterName})

	// アサーション
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// 最初のサービスを検証
	assert.Equal(t, "web-service", result[0].ServiceName)
	assert.Equal(t, "test-cluster", result[0].ClusterName)
	assert.Equal(t, "web-task:1", result[0].TaskDefinition)
	assert.Equal(t, int32(2), result[0].DesiredCount)
	assert.Equal(t, int32(2), result[0].RunningCount)
	assert.Equal(t, "ACTIVE", result[0].Status)

	// 2番目のサービスを検証
	assert.Equal(t, "api-service", result[1].ServiceName)
	assert.Equal(t, "test-cluster", result[1].ClusterName)

	mockClient.AssertExpectations(t)
}

func TestScanner_ScanServices_MultipleClusters(t *testing.T) {
	mockClient := new(MockECSClient)
	scanner := scanner.NewScanner(mockClient)

	ctx := context.Background()
	clusters := []string{"cluster1", "cluster2"}

	// cluster1のモック設定
	mockClient.On("ListServices", ctx, &ecs.ListServicesInput{
		Cluster: &clusters[0],
	}).Return(
		&ecs.ListServicesOutput{
			ServiceArns: []string{"arn:aws:ecs:us-west-2:123456789012:service/cluster1/service1"},
		}, nil)

	mockClient.On("DescribeServices", ctx, &ecs.DescribeServicesInput{
		Cluster:  &clusters[0],
		Services: []string{"arn:aws:ecs:us-west-2:123456789012:service/cluster1/service1"},
	}).Return(
		&ecs.DescribeServicesOutput{
			Services: []types.Service{
				{
					ServiceName:    stringPtr("service1"),
					ServiceArn:     stringPtr("arn:aws:ecs:us-west-2:123456789012:service/cluster1/service1"),
					ClusterArn:     stringPtr("arn:aws:ecs:us-west-2:123456789012:cluster/cluster1"),
					TaskDefinition: stringPtr("task1:1"),
					DesiredCount:   1,
					RunningCount:   1,
					Status:         stringPtr("ACTIVE"),
				},
			},
		}, nil)

	// cluster2のモック設定
	mockClient.On("ListServices", ctx, &ecs.ListServicesInput{
		Cluster: &clusters[1],
	}).Return(
		&ecs.ListServicesOutput{
			ServiceArns: []string{"arn:aws:ecs:us-west-2:123456789012:service/cluster2/service2"},
		}, nil)

	mockClient.On("DescribeServices", ctx, &ecs.DescribeServicesInput{
		Cluster:  &clusters[1],
		Services: []string{"arn:aws:ecs:us-west-2:123456789012:service/cluster2/service2"},
	}).Return(
		&ecs.DescribeServicesOutput{
			Services: []types.Service{
				{
					ServiceName:    stringPtr("service2"),
					ServiceArn:     stringPtr("arn:aws:ecs:us-west-2:123456789012:service/cluster2/service2"),
					ClusterArn:     stringPtr("arn:aws:ecs:us-west-2:123456789012:cluster/cluster2"),
					TaskDefinition: stringPtr("task2:1"),
					DesiredCount:   2,
					RunningCount:   2,
					Status:         stringPtr("ACTIVE"),
				},
			},
		}, nil)

	// テスト実行
	result, err := scanner.ScanServices(ctx, clusters)

	// アサーション
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// 各クラスターからサービスが取得されることを確認
	assert.Equal(t, "service1", result[0].ServiceName)
	assert.Equal(t, "cluster1", result[0].ClusterName)
	assert.Equal(t, "service2", result[1].ServiceName)
	assert.Equal(t, "cluster2", result[1].ClusterName)

	mockClient.AssertExpectations(t)
}

func TestScanner_DiscoverClusters(t *testing.T) {
	mockClient := new(MockECSClient)
	scanner := scanner.NewScanner(mockClient)

	ctx := context.Background()

	// モックの設定
	mockClient.On("ListClusters", ctx, &ecs.ListClustersInput{}).Return(
		&ecs.ListClustersOutput{
			ClusterArns: []string{
				"arn:aws:ecs:us-west-2:123456789012:cluster/cluster1",
				"arn:aws:ecs:us-west-2:123456789012:cluster/cluster2",
				"arn:aws:ecs:us-west-2:123456789012:cluster/cluster3",
			},
		}, nil)

	// テスト実行
	clusters, err := scanner.DiscoverClusters(ctx)

	// アサーション
	assert.NoError(t, err)
	assert.Len(t, clusters, 3)
	assert.Equal(t, "cluster1", clusters[0])
	assert.Equal(t, "cluster2", clusters[1])
	assert.Equal(t, "cluster3", clusters[2])

	mockClient.AssertExpectations(t)
}

func TestScanner_ScanServices_EmptyCluster(t *testing.T) {
	mockClient := new(MockECSClient)
	scanner := scanner.NewScanner(mockClient)

	ctx := context.Background()
	clusterName := "empty-cluster"

	// モックの設定 - 空のサービス一覧
	mockClient.On("ListServices", ctx, &ecs.ListServicesInput{
		Cluster: &clusterName,
	}).Return(
		&ecs.ListServicesOutput{
			ServiceArns: []string{},
		}, nil)

	// テスト実行
	result, err := scanner.ScanServices(ctx, []string{clusterName})

	// アサーション
	assert.NoError(t, err)
	assert.Len(t, result, 0)

	mockClient.AssertExpectations(t)
}

// ヘルパー関数
func stringPtr(s string) *string {
	return &s
}
