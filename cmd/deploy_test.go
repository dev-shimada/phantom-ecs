package cmd_test

import (
	"context"
	"testing"

	"github.com/dev-shimada/phantom-ecs/cmd"
	"github.com/dev-shimada/phantom-ecs/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeployer はDeployerのモック
type MockDeployer struct {
	mock.Mock
}

func (m *MockDeployer) DeployService(ctx context.Context, inspectionResult *models.InspectionResult, targetCluster, newServiceName string, dryRun bool) (*models.DeploymentResult, error) {
	args := m.Called(ctx, inspectionResult, targetCluster, newServiceName, dryRun)
	return args.Get(0).(*models.DeploymentResult), args.Error(1)
}

// MockInspectorForDeploy はDeploy用のInspectorモック
type MockInspectorForDeploy struct {
	mock.Mock
}

func (m *MockInspectorForDeploy) InspectService(ctx context.Context, serviceName, clusterName string) (*models.InspectionResult, error) {
	args := m.Called(ctx, serviceName, clusterName)
	return args.Get(0).(*models.InspectionResult), args.Error(1)
}

func TestDeployCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError bool
		setupMocks    func(*MockDeployer, *MockInspectorForDeploy)
	}{
		{
			name:          "基本的なデプロイ（ドライラン）",
			args:          []string{"deploy", "source-service", "--from-cluster", "source-cluster", "--target-cluster", "target-cluster", "--dry-run"},
			expectedError: false,
			setupMocks: func(mockDeployer *MockDeployer, mockInspector *MockInspectorForDeploy) {
				inspectionResult := &models.InspectionResult{
					Service: models.ECSService{
						ServiceName:    "source-service",
						ClusterName:    "source-cluster",
						Status:         "ACTIVE",
						TaskDefinition: "source-task-def:1",
						DesiredCount:   1,
						RunningCount:   1,
						LaunchType:     "FARGATE",
					},
					TaskDefinition: models.ECSTaskDefinition{
						Family:      "source-task-def",
						Revision:    1,
						Status:      "ACTIVE",
						CPU:         "256",
						Memory:      "512",
						NetworkMode: "awsvpc",
					},
				}
				mockInspector.On("InspectService", mock.Anything, "source-service", "source-cluster").Return(inspectionResult, nil)
				mockDeployer.On("DeployService", mock.Anything, inspectionResult, "target-cluster", "source-service", true).Return(&models.DeploymentResult{
					ServiceName: "source-service",
					ClusterName: "target-cluster",
					Success:     true,
					DryRun:      true,
					Operations:  []string{"Register task definition: source-task-def-copy", "Create service: source-service in cluster target-cluster"},
				}, nil)
			},
		},
		{
			name:          "実際のデプロイ",
			args:          []string{"deploy", "prod-service", "--from-cluster", "prod-cluster", "--target-cluster", "staging-cluster", "--new-service-name", "staging-prod-service"},
			expectedError: false,
			setupMocks: func(mockDeployer *MockDeployer, mockInspector *MockInspectorForDeploy) {
				inspectionResult := &models.InspectionResult{
					Service: models.ECSService{
						ServiceName:    "prod-service",
						ClusterName:    "prod-cluster",
						Status:         "ACTIVE",
						TaskDefinition: "prod-task-def:5",
						DesiredCount:   3,
						RunningCount:   3,
						LaunchType:     "EC2",
					},
					TaskDefinition: models.ECSTaskDefinition{
						Family:      "prod-task-def",
						Revision:    5,
						Status:      "ACTIVE",
						CPU:         "1024",
						Memory:      "2048",
						NetworkMode: "bridge",
					},
				}
				mockInspector.On("InspectService", mock.Anything, "prod-service", "prod-cluster").Return(inspectionResult, nil)
				mockDeployer.On("DeployService", mock.Anything, inspectionResult, "staging-cluster", "staging-prod-service", false).Return(&models.DeploymentResult{
					ServiceName:       "staging-prod-service",
					ClusterName:       "staging-cluster",
					TaskDefinitionArn: "arn:aws:ecs:us-east-1:123456789012:task-definition/prod-task-def-copy:1",
					Success:           true,
					DryRun:            false,
				}, nil)
			},
		},
		{
			name:          "サービス名未指定エラー",
			args:          []string{"deploy"},
			expectedError: true,
			setupMocks: func(mockDeployer *MockDeployer, mockInspector *MockInspectorForDeploy) {
				// エラーの場合はモックを設定しない
			},
		},
		{
			name:          "ソースクラスター未指定エラー",
			args:          []string{"deploy", "test-service"},
			expectedError: true,
			setupMocks: func(mockDeployer *MockDeployer, mockInspector *MockInspectorForDeploy) {
				// エラーの場合はモックを設定しない
			},
		},
		{
			name:          "ターゲットクラスター未指定エラー",
			args:          []string{"deploy", "test-service", "--from-cluster", "source-cluster"},
			expectedError: true,
			setupMocks: func(mockDeployer *MockDeployer, mockInspector *MockInspectorForDeploy) {
				// エラーの場合はモックを設定しない
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDeployer := &MockDeployer{}
			mockInspector := &MockInspectorForDeploy{}
			tt.setupMocks(mockDeployer, mockInspector)

			cmd := cmd.NewDeployCommand(mockDeployer, mockInspector)
			cmd.SetArgs(tt.args[1:]) // "deploy"を除く

			err := cmd.Execute()
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDeployer.AssertExpectations(t)
			mockInspector.AssertExpectations(t)
		})
	}
}

func TestDeployCommandFlags(t *testing.T) {
	mockDeployer := &MockDeployer{}
	mockInspector := &MockInspectorForDeploy{}
	cmd := cmd.NewDeployCommand(mockDeployer, mockInspector)

	// フラグの存在確認
	assert.NotNil(t, cmd.Flags().Lookup("from-cluster"))
	assert.NotNil(t, cmd.Flags().Lookup("target-cluster"))
	assert.NotNil(t, cmd.Flags().Lookup("new-service-name"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("region"))
	assert.NotNil(t, cmd.Flags().Lookup("profile"))
	assert.NotNil(t, cmd.Flags().Lookup("output"))
}

func TestDeployCommandHelp(t *testing.T) {
	mockDeployer := &MockDeployer{}
	mockInspector := &MockInspectorForDeploy{}
	cmd := cmd.NewDeployCommand(mockDeployer, mockInspector)

	// コマンドの基本情報確認
	assert.Equal(t, "deploy <service-name>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
}
