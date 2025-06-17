package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromYAMLFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "phantom-ecs.yaml")

	yamlContent := `
profiles:
  default:
    region: us-west-2
    output_format: json
  production:
    region: ap-northeast-1
    output_format: table
  development:
    region: us-east-1
    output_format: yaml

batch:
  max_concurrency: 5
  retry_attempts: 3
  retry_delay: 2s
  show_progress: true

logging:
  level: info
  format: json
  filename: /var/log/phantom-ecs.log
  max_size: 100
  max_age: 30
  max_backups: 10
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	config, err := LoadFromFile(configFile, "default")
	require.NoError(t, err)

	assert.Equal(t, "us-west-2", config.Region)
	assert.Equal(t, "json", config.OutputFormat)
	assert.Equal(t, 5, config.Batch.MaxConcurrency)
	assert.Equal(t, "info", config.Logging.Level)
}

func TestLoadFromYAMLFile_DifferentProfile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "phantom-ecs.yaml")

	yamlContent := `
profiles:
  production:
    region: ap-northeast-1
    output_format: table
    aws_profile: prod-profile
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	config, err := LoadFromFile(configFile, "production")
	require.NoError(t, err)

	assert.Equal(t, "ap-northeast-1", config.Region)
	assert.Equal(t, "table", config.OutputFormat)
	assert.Equal(t, "prod-profile", config.Profile)
}

func TestLoadFromYAMLFile_ProfileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "phantom-ecs.yaml")

	yamlContent := `
profiles:
  default:
    region: us-west-2
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	_, err = LoadFromFile(configFile, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "プロファイル 'nonexistent' が見つかりません")
}

func TestLoadFromYAMLFile_FileNotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/file.yaml", "default")
	assert.Error(t, err)
}

func TestEnhancedConfig_WithEnvironmentVariables(t *testing.T) {
	// 環境変数を設定
	os.Setenv("PHANTOM_ECS_REGION", "eu-west-1")
	os.Setenv("PHANTOM_ECS_OUTPUT_FORMAT", "yaml")
	os.Setenv("PHANTOM_ECS_LOG_LEVEL", "debug")
	os.Setenv("PHANTOM_ECS_BATCH_MAX_CONCURRENCY", "10")
	defer func() {
		os.Unsetenv("PHANTOM_ECS_REGION")
		os.Unsetenv("PHANTOM_ECS_OUTPUT_FORMAT")
		os.Unsetenv("PHANTOM_ECS_LOG_LEVEL")
		os.Unsetenv("PHANTOM_ECS_BATCH_MAX_CONCURRENCY")
	}()

	config := NewEnhancedConfigFromEnvironment()

	assert.Equal(t, "eu-west-1", config.Region)
	assert.Equal(t, "yaml", config.OutputFormat)
	assert.Equal(t, "debug", config.Logging.Level)
	assert.Equal(t, 10, config.Batch.MaxConcurrency)
}

func TestEnhancedConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      *EnhancedConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "有効な設定",
			config: &EnhancedConfig{
				Config: Config{
					Region:       "us-east-1",
					OutputFormat: "json",
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Batch: BatchConfig{
					MaxConcurrency: 3,
					RetryAttempts:  3,
				},
			},
			expectError: false,
		},
		{
			name: "無効なリージョン",
			config: &EnhancedConfig{
				Config: Config{
					Region:       "invalid-region",
					OutputFormat: "json",
				},
			},
			expectError: true,
			errorMsg:    "invalid AWS region",
		},
		{
			name: "無効な出力フォーマット",
			config: &EnhancedConfig{
				Config: Config{
					Region:       "us-east-1",
					OutputFormat: "invalid-format",
				},
			},
			expectError: true,
			errorMsg:    "無効な出力フォーマット",
		},
		{
			name: "無効なログレベル",
			config: &EnhancedConfig{
				Config: Config{
					Region:       "us-east-1",
					OutputFormat: "json",
				},
				Logging: LoggingConfig{
					Level:  "invalid-level",
					Format: "json",
				},
			},
			expectError: true,
			errorMsg:    "無効なログレベル",
		},
		{
			name: "無効な同時実行数",
			config: &EnhancedConfig{
				Config: Config{
					Region:       "us-east-1",
					OutputFormat: "json",
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Batch: BatchConfig{
					MaxConcurrency: 0,
				},
			},
			expectError: true,
			errorMsg:    "同時実行数は1以上である必要があります",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDefaultEnhancedConfig(t *testing.T) {
	config := GetDefaultEnhancedConfig()

	assert.Equal(t, DefaultRegion, config.Region)
	assert.Equal(t, DefaultOutputFormat, config.OutputFormat)
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "text", config.Logging.Format)
	assert.Equal(t, 3, config.Batch.MaxConcurrency)
	assert.Equal(t, 3, config.Batch.RetryAttempts)
}

func TestMergeWithEnvironment(t *testing.T) {
	config := &EnhancedConfig{
		Config: Config{
			Region:       "us-east-1",
			OutputFormat: "json",
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}

	// 環境変数を設定
	os.Setenv("PHANTOM_ECS_REGION", "eu-west-1")
	os.Setenv("PHANTOM_ECS_LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("PHANTOM_ECS_REGION")
		os.Unsetenv("PHANTOM_ECS_LOG_LEVEL")
	}()

	config.MergeWithEnvironment()

	assert.Equal(t, "eu-west-1", config.Region)
	assert.Equal(t, "json", config.OutputFormat) // 環境変数で上書きされない
	assert.Equal(t, "debug", config.Logging.Level)
}

func TestSaveToFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	config := GetDefaultEnhancedConfig()
	config.Region = "ap-northeast-1"
	config.OutputFormat = "yaml"

	err := config.SaveToFile(configFile)
	require.NoError(t, err)

	// ファイルが作成されたことを確認
	assert.FileExists(t, configFile)

	// ファイルから読み込んで設定が正しく保存されているか確認
	loadedConfig, err := LoadFromFile(configFile, "default")
	require.NoError(t, err)

	assert.Equal(t, "ap-northeast-1", loadedConfig.Region)
	assert.Equal(t, "yaml", loadedConfig.OutputFormat)
}
