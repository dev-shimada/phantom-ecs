package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		region      string
		profile     string
		expectError bool
	}{
		{
			name:        "valid configuration",
			region:      "us-east-1",
			profile:     "",
			expectError: false,
		},
		{
			name:        "invalid profile should return error",
			region:      "us-west-2",
			profile:     "nonexistent-profile",
			expectError: true,
		},
		{
			name:        "empty region should use default",
			region:      "",
			profile:     "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(context.Background(), tt.region, tt.profile)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_GetECSClient(t *testing.T) {
	client, err := NewClient(context.Background(), "us-east-1", "")
	require.NoError(t, err)
	require.NotNil(t, client)

	ecsClient := client.GetECSClient()
	assert.NotNil(t, ecsClient)
}

func TestClient_GetRegion(t *testing.T) {
	tests := []struct {
		name           string
		region         string
		expectedRegion string
	}{
		{
			name:           "explicit region",
			region:         "us-west-2",
			expectedRegion: "us-west-2",
		},
		{
			name:           "default region when empty",
			region:         "",
			expectedRegion: "us-east-1", // デフォルトリージョン
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(context.Background(), tt.region, "")
			require.NoError(t, err)

			actualRegion := client.GetRegion()
			assert.Equal(t, tt.expectedRegion, actualRegion)
		})
	}
}
