package cmd

import (
	"context"
	"testing"

	"github.com/dev-shimada/phantom-ecs/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockScanner はScannerのモック
type MockScanner struct {
	mock.Mock
}

func (m *MockScanner) ScanServices(ctx context.Context, clusterNames []string) ([]models.ECSService, error) {
	args := m.Called(ctx, clusterNames)
	return args.Get(0).([]models.ECSService), args.Error(1)
}

func (m *MockScanner) DiscoverClusters(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func TestScanCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError bool
		setupMock     func(*MockScanner)
	}{
		{
			name:          "デフォルト設定でスキャン",
			args:          []string{"scan"},
			expectedError: false,
			setupMock: func(m *MockScanner) {
				m.On("DiscoverClusters", mock.Anything).Return([]string{"test-cluster"}, nil)
				m.On("ScanServices", mock.Anything, []string{"test-cluster"}).Return([]models.ECSService{
					{
						ServiceName:    "test-service",
						ClusterName:    "test-cluster",
						Status:         "ACTIVE",
						TaskDefinition: "test-task-def:1",
						DesiredCount:   1,
						RunningCount:   1,
						LaunchType:     "FARGATE",
					},
				}, nil)
			},
		},
		{
			name:          "特定リージョンでスキャン",
			args:          []string{"scan", "--region", "us-west-2"},
			expectedError: false,
			setupMock: func(m *MockScanner) {
				m.On("DiscoverClusters", mock.Anything).Return([]string{"west-cluster"}, nil)
				m.On("ScanServices", mock.Anything, []string{"west-cluster"}).Return([]models.ECSService{}, nil)
			},
		},
		{
			name:          "JSON出力形式",
			args:          []string{"scan", "--output", "json"},
			expectedError: false,
			setupMock: func(m *MockScanner) {
				m.On("DiscoverClusters", mock.Anything).Return([]string{"test-cluster"}, nil)
				m.On("ScanServices", mock.Anything, []string{"test-cluster"}).Return([]models.ECSService{
					{
						ServiceName:    "json-service",
						ClusterName:    "test-cluster",
						Status:         "ACTIVE",
						TaskDefinition: "json-task-def:1",
						DesiredCount:   2,
						RunningCount:   2,
						LaunchType:     "EC2",
					},
				}, nil)
			},
		},
		{
			name:          "無効な出力形式",
			args:          []string{"scan", "--output", "invalid"},
			expectedError: true,
			setupMock: func(m *MockScanner) {
				// エラーの場合はモックを設定しない
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockScanner := &MockScanner{}
			tt.setupMock(mockScanner)

			cmd := NewScanCommand(mockScanner)
			cmd.SetArgs(tt.args[1:]) // "scan"を除く

			err := cmd.Execute()
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockScanner.AssertExpectations(t)
		})
	}
}

func TestScanCommandFlags(t *testing.T) {
	mockScanner := &MockScanner{}
	cmd := NewScanCommand(mockScanner)

	// フラグの存在確認
	assert.NotNil(t, cmd.Flags().Lookup("region"))
	assert.NotNil(t, cmd.Flags().Lookup("profile"))
	assert.NotNil(t, cmd.Flags().Lookup("output"))
}

func TestScanCommandHelp(t *testing.T) {
	mockScanner := &MockScanner{}
	cmd := NewScanCommand(mockScanner)

	// コマンドの基本情報確認
	assert.Equal(t, "scan", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
}
