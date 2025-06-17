package cmd

import (
	"context"
	"testing"

	"github.com/dev-shimada/phantom-ecs/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockInspector はInspectorのモック
type MockInspector struct {
	mock.Mock
}

func (m *MockInspector) InspectService(ctx context.Context, serviceName, clusterName string) (*models.InspectionResult, error) {
	args := m.Called(ctx, serviceName, clusterName)
	return args.Get(0).(*models.InspectionResult), args.Error(1)
}

func TestInspectCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError bool
		setupMock     func(*MockInspector)
	}{
		{
			name:          "基本的なサービス検査",
			args:          []string{"inspect", "test-service", "--cluster", "test-cluster"},
			expectedError: false,
			setupMock: func(m *MockInspector) {
				m.On("InspectService", mock.Anything, "test-service", "test-cluster").Return(&models.InspectionResult{
					Service: models.ECSService{
						ServiceName:    "test-service",
						ClusterName:    "test-cluster",
						Status:         "ACTIVE",
						TaskDefinition: "test-task-def:1",
						DesiredCount:   1,
						RunningCount:   1,
						LaunchType:     "FARGATE",
					},
					TaskDefinition: models.ECSTaskDefinition{
						Family:      "test-task-def",
						Revision:    1,
						Status:      "ACTIVE",
						CPU:         "256",
						Memory:      "512",
						NetworkMode: "awsvpc",
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
							Description: "Enable ECS Service Auto Scaling for better resource utilization",
							Priority:    "medium",
							Action:      "Configure Auto Scaling policies based on CPU and memory utilization",
						},
					},
				}, nil)
			},
		},
		{
			name:          "JSON出力形式",
			args:          []string{"inspect", "json-service", "--cluster", "json-cluster", "--output", "json"},
			expectedError: false,
			setupMock: func(m *MockInspector) {
				m.On("InspectService", mock.Anything, "json-service", "json-cluster").Return(&models.InspectionResult{
					Service: models.ECSService{
						ServiceName:    "json-service",
						ClusterName:    "json-cluster",
						Status:         "ACTIVE",
						TaskDefinition: "json-task-def:2",
						DesiredCount:   2,
						RunningCount:   2,
						LaunchType:     "EC2",
					},
					TaskDefinition: models.ECSTaskDefinition{
						Family:      "json-task-def",
						Revision:    2,
						Status:      "ACTIVE",
						CPU:         "512",
						Memory:      "1024",
						NetworkMode: "bridge",
					},
					Recommendations: []models.Recommendation{},
				}, nil)
			},
		},
		{
			name:          "サービス名未指定エラー",
			args:          []string{"inspect"},
			expectedError: true,
			setupMock: func(m *MockInspector) {
				// エラーの場合はモックを設定しない
			},
		},
		{
			name:          "クラスター名未指定エラー",
			args:          []string{"inspect", "test-service"},
			expectedError: true,
			setupMock: func(m *MockInspector) {
				// エラーの場合はモックを設定しない
			},
		},
		{
			name:          "無効な出力形式",
			args:          []string{"inspect", "test-service", "--cluster", "test-cluster", "--output", "invalid"},
			expectedError: true,
			setupMock: func(m *MockInspector) {
				// エラーの場合はモックを設定しない
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInspector := &MockInspector{}
			tt.setupMock(mockInspector)

			cmd := NewInspectCommand(mockInspector)
			cmd.SetArgs(tt.args[1:]) // "inspect"を除く

			err := cmd.Execute()
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockInspector.AssertExpectations(t)
		})
	}
}

func TestInspectCommandFlags(t *testing.T) {
	mockInspector := &MockInspector{}
	cmd := NewInspectCommand(mockInspector)

	// フラグの存在確認
	assert.NotNil(t, cmd.Flags().Lookup("cluster"))
	assert.NotNil(t, cmd.Flags().Lookup("region"))
	assert.NotNil(t, cmd.Flags().Lookup("profile"))
	assert.NotNil(t, cmd.Flags().Lookup("output"))
}

func TestInspectCommandHelp(t *testing.T) {
	mockInspector := &MockInspector{}
	cmd := NewInspectCommand(mockInspector)

	// コマンドの基本情報確認
	assert.Equal(t, "inspect <service-name>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
}

func TestInspectCommandArgs(t *testing.T) {
	mockInspector := &MockInspector{}
	cmd := NewInspectCommand(mockInspector)

	// 引数の検証確認
	assert.NotNil(t, cmd.Args)
}
