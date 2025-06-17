package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestECSService_JsonMarshaling(t *testing.T) {
	service := &ECSService{
		ServiceName:    "test-service",
		ClusterName:    "test-cluster",
		Status:         "ACTIVE",
		TaskDefinition: "test-task-def:1",
		DesiredCount:   2,
		RunningCount:   2,
		CreatedAt:      time.Now(),
		LaunchType:     "FARGATE",
	}

	// JSON マーシャリングのテスト
	jsonData, err := json.Marshal(service)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// JSON アンマーシャリングのテスト
	var unmarshaledService ECSService
	err = json.Unmarshal(jsonData, &unmarshaledService)
	require.NoError(t, err)
	assert.Equal(t, service.ServiceName, unmarshaledService.ServiceName)
	assert.Equal(t, service.ClusterName, unmarshaledService.ClusterName)
	assert.Equal(t, service.Status, unmarshaledService.Status)
}

func TestECSService_IsHealthy(t *testing.T) {
	tests := []struct {
		name          string
		service       *ECSService
		expectHealthy bool
	}{
		{
			name: "healthy service",
			service: &ECSService{
				Status:       "ACTIVE",
				DesiredCount: 2,
				RunningCount: 2,
			},
			expectHealthy: true,
		},
		{
			name: "unhealthy service - wrong status",
			service: &ECSService{
				Status:       "INACTIVE",
				DesiredCount: 2,
				RunningCount: 2,
			},
			expectHealthy: false,
		},
		{
			name: "unhealthy service - count mismatch",
			service: &ECSService{
				Status:       "ACTIVE",
				DesiredCount: 2,
				RunningCount: 1,
			},
			expectHealthy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.service.IsHealthy()
			assert.Equal(t, tt.expectHealthy, result)
		})
	}
}

func TestECSTaskDefinition_GetFamilyAndRevision(t *testing.T) {
	tests := []struct {
		name             string
		taskDef          *ECSTaskDefinition
		expectedFamily   string
		expectedRevision int
	}{
		{
			name: "valid task definition arn",
			taskDef: &ECSTaskDefinition{
				TaskDefinitionArn: "arn:aws:ecs:us-east-1:123456789012:task-definition/my-task:5",
			},
			expectedFamily:   "my-task",
			expectedRevision: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			family, revision := tt.taskDef.GetFamilyAndRevision()
			assert.Equal(t, tt.expectedFamily, family)
			assert.Equal(t, tt.expectedRevision, revision)
		})
	}
}
