package phantomecs_test

import (
	"context"
	"testing"

	"github.com/dev-shimada/phantom-ecs/pkg/phantomecs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPhantomECSClient(t *testing.T) {
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
			name:        "empty region should use default",
			region:      "",
			profile:     "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := phantomecs.NewPhantomECSClient(context.Background(), tt.region, tt.profile)

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

func TestPhantomECSClient_GetConfig(t *testing.T) {
	client, err := phantomecs.NewPhantomECSClient(context.Background(), "us-west-2", "")
	require.NoError(t, err)
	require.NotNil(t, client)

	config := client.GetConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "us-west-2", config.GetRegion())
	assert.Equal(t, "", config.GetProfile())
}

func TestPhantomECSClient_GetECSService(t *testing.T) {
	client, err := phantomecs.NewPhantomECSClient(context.Background(), "us-east-1", "")
	require.NoError(t, err)
	require.NotNil(t, client)

	ecsService := client.GetECSService()
	assert.NotNil(t, ecsService)
}
