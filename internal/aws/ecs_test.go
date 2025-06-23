package aws_test

import (
	"context"
	"testing"

	"github.com/dev-shimada/phantom-ecs/internal/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockECSClient テスト用のモックECSクライアント
type MockECSClient struct {
	listServicesOutput     []string
	describeServicesOutput map[string]interface{}
	describeTaskDefOutput  map[string]interface{}
	shouldReturnError      bool
}

func (m *MockECSClient) ListServices(ctx context.Context, clusterName string) ([]string, error) {
	if m.shouldReturnError {
		return nil, assert.AnError
	}
	return m.listServicesOutput, nil
}

func (m *MockECSClient) DescribeServices(ctx context.Context, clusterName string, serviceNames []string) (map[string]interface{}, error) {
	if m.shouldReturnError {
		return nil, assert.AnError
	}
	return m.describeServicesOutput, nil
}

func (m *MockECSClient) DescribeTaskDefinition(ctx context.Context, taskDefArn string) (map[string]interface{}, error) {
	if m.shouldReturnError {
		return nil, assert.AnError
	}
	return m.describeTaskDefOutput, nil
}

func TestNewECSService(t *testing.T) {
	client, err := aws.NewClient(context.Background(), "us-east-1", "")
	require.NoError(t, err)
	require.NotNil(t, client)

	ecsService := aws.NewECSService(client)
	assert.NotNil(t, ecsService)
	assert.Equal(t, client, ecsService.GetClient())
}

func TestECSService_ListServices(t *testing.T) {
	mockClient := &MockECSClient{
		listServicesOutput: []string{"service1", "service2", "service3"},
		shouldReturnError:  false,
	}

	tests := []struct {
		name        string
		clusterName string
		expectError bool
		expectedLen int
	}{
		{
			name:        "successful list services",
			clusterName: "test-cluster",
			expectError: false,
			expectedLen: 3,
		},
		{
			name:        "empty cluster name",
			clusterName: "",
			expectError: false,
			expectedLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services, err := mockClient.ListServices(context.Background(), tt.clusterName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, services)
			} else {
				assert.NoError(t, err)
				assert.Len(t, services, tt.expectedLen)
			}
		})
	}
}

func TestECSService_DescribeServices(t *testing.T) {
	mockClient := &MockECSClient{
		describeServicesOutput: map[string]interface{}{
			"services": []map[string]interface{}{
				{
					"serviceName":  "test-service",
					"status":       "ACTIVE",
					"desiredCount": 2,
					"runningCount": 2,
				},
			},
		},
		shouldReturnError: false,
	}

	serviceNames := []string{"test-service"}
	result, err := mockClient.DescribeServices(context.Background(), "test-cluster", serviceNames)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "services")
}

func TestECSService_DescribeTaskDefinition(t *testing.T) {
	mockClient := &MockECSClient{
		describeTaskDefOutput: map[string]interface{}{
			"taskDefinition": map[string]interface{}{
				"family":   "test-task-family",
				"revision": 1,
				"status":   "ACTIVE",
			},
		},
		shouldReturnError: false,
	}

	result, err := mockClient.DescribeTaskDefinition(context.Background(), "test-task-arn")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "taskDefinition")
}
