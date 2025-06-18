package config_test

import (
	"os"
	"testing"

	"github.com/dev-shimada/phantom-ecs/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
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
			name:        "valid configuration with profile",
			region:      "us-west-2",
			profile:     "test-profile",
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
			config := config.NewConfig(tt.region, tt.profile)

			if tt.expectError {
				assert.Nil(t, config)
			} else {
				assert.NotNil(t, config)
				if tt.region == "" {
					assert.Equal(t, "us-east-1", config.GetRegion()) // デフォルトリージョン
				} else {
					assert.Equal(t, tt.region, config.GetRegion())
				}
				assert.Equal(t, tt.profile, config.GetProfile())
			}
		})
	}
}

func TestConfig_FromEnvironment(t *testing.T) {
	// 環境変数をセットアップ
	os.Setenv("AWS_REGION", "eu-west-1")
	os.Setenv("AWS_PROFILE", "test-env-profile")
	defer func() {
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_PROFILE")
	}()

	config := config.NewConfigFromEnvironment()
	require.NotNil(t, config)

	assert.Equal(t, "eu-west-1", config.GetRegion())
	assert.Equal(t, "test-env-profile", config.GetProfile())
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name: "valid configuration",
			config: &config.Config{
				Region:  "us-east-1",
				Profile: "",
			},
			expectError: false,
		},
		{
			name: "invalid region",
			config: &config.Config{
				Region:  "invalid-region",
				Profile: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_SetOutputFormat(t *testing.T) {
	config := config.NewConfig("us-east-1", "")
	require.NotNil(t, config)

	// デフォルトフォーマットの確認
	assert.Equal(t, "table", config.GetOutputFormat())

	// JSONフォーマットの設定
	config.SetOutputFormat("json")
	assert.Equal(t, "json", config.GetOutputFormat())

	// YAMLフォーマットの設定
	config.SetOutputFormat("yaml")
	assert.Equal(t, "yaml", config.GetOutputFormat())
}
