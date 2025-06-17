package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// EnhancedConfig は拡張された設定構造体
type EnhancedConfig struct {
	Config  `yaml:",inline"`
	Logging LoggingConfig `yaml:"logging"`
	Batch   BatchConfig   `yaml:"batch"`
}

// LoggingConfig はロギング設定
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
}

// BatchConfig はバッチ処理設定
type BatchConfig struct {
	MaxConcurrency int           `yaml:"max_concurrency"`
	RetryAttempts  int           `yaml:"retry_attempts"`
	RetryDelay     time.Duration `yaml:"retry_delay"`
	ShowProgress   bool          `yaml:"show_progress"`
}

// ProfileConfig はプロファイル別設定
type ProfileConfig struct {
	Region       string `yaml:"region"`
	OutputFormat string `yaml:"output_format"`
	AWSProfile   string `yaml:"aws_profile"`
}

// FileConfig はYAMLファイルの構造
type FileConfig struct {
	Profiles map[string]ProfileConfig `yaml:"profiles"`
	Logging  LoggingConfig            `yaml:"logging"`
	Batch    BatchConfig              `yaml:"batch"`
}

// LoadFromFile はYAMLファイルから設定を読み込む
func LoadFromFile(filename, profileName string) (*EnhancedConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	var fileConfig FileConfig
	if err := yaml.Unmarshal(data, &fileConfig); err != nil {
		return nil, fmt.Errorf("YAML解析に失敗しました: %w", err)
	}

	profile, exists := fileConfig.Profiles[profileName]
	if !exists {
		return nil, fmt.Errorf("プロファイル '%s' が見つかりません", profileName)
	}

	config := &EnhancedConfig{
		Config: Config{
			Region:       profile.Region,
			Profile:      profile.AWSProfile,
			OutputFormat: profile.OutputFormat,
		},
		Logging: fileConfig.Logging,
		Batch:   fileConfig.Batch,
	}

	// デフォルト値の設定
	config.setDefaults()

	return config, nil
}

// NewEnhancedConfigFromEnvironment は環境変数から拡張設定を作成する
func NewEnhancedConfigFromEnvironment() *EnhancedConfig {
	config := &EnhancedConfig{
		Config: Config{
			Region:       getEnvOrDefault("PHANTOM_ECS_REGION", DefaultRegion),
			Profile:      getEnvOrDefault("PHANTOM_ECS_PROFILE", ""),
			OutputFormat: getEnvOrDefault("PHANTOM_ECS_OUTPUT_FORMAT", DefaultOutputFormat),
		},
		Logging: LoggingConfig{
			Level:      getEnvOrDefault("PHANTOM_ECS_LOG_LEVEL", "info"),
			Format:     getEnvOrDefault("PHANTOM_ECS_LOG_FORMAT", "text"),
			Filename:   getEnvOrDefault("PHANTOM_ECS_LOG_FILE", ""),
			MaxSize:    getEnvIntOrDefault("PHANTOM_ECS_LOG_MAX_SIZE", 100),
			MaxAge:     getEnvIntOrDefault("PHANTOM_ECS_LOG_MAX_AGE", 30),
			MaxBackups: getEnvIntOrDefault("PHANTOM_ECS_LOG_MAX_BACKUPS", 10),
		},
		Batch: BatchConfig{
			MaxConcurrency: getEnvIntOrDefault("PHANTOM_ECS_BATCH_MAX_CONCURRENCY", 3),
			RetryAttempts:  getEnvIntOrDefault("PHANTOM_ECS_BATCH_RETRY_ATTEMPTS", 3),
			RetryDelay:     getEnvDurationOrDefault("PHANTOM_ECS_BATCH_RETRY_DELAY", time.Second*2),
			ShowProgress:   getEnvBoolOrDefault("PHANTOM_ECS_BATCH_SHOW_PROGRESS", true),
		},
	}

	return config
}

// GetDefaultEnhancedConfig はデフォルトの拡張設定を返す
func GetDefaultEnhancedConfig() *EnhancedConfig {
	return &EnhancedConfig{
		Config: Config{
			Region:       DefaultRegion,
			Profile:      "",
			OutputFormat: DefaultOutputFormat,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			Filename:   "",
			MaxSize:    100,
			MaxAge:     30,
			MaxBackups: 10,
		},
		Batch: BatchConfig{
			MaxConcurrency: 3,
			RetryAttempts:  3,
			RetryDelay:     time.Second * 2,
			ShowProgress:   true,
		},
	}
}

// setDefaults はデフォルト値を設定する
func (c *EnhancedConfig) setDefaults() {
	if c.Region == "" {
		c.Region = DefaultRegion
	}
	if c.OutputFormat == "" {
		c.OutputFormat = DefaultOutputFormat
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "text"
	}
	if c.Logging.MaxSize == 0 {
		c.Logging.MaxSize = 100
	}
	if c.Logging.MaxAge == 0 {
		c.Logging.MaxAge = 30
	}
	if c.Logging.MaxBackups == 0 {
		c.Logging.MaxBackups = 10
	}
	if c.Batch.MaxConcurrency == 0 {
		c.Batch.MaxConcurrency = 3
	}
	if c.Batch.RetryAttempts == 0 {
		c.Batch.RetryAttempts = 3
	}
	if c.Batch.RetryDelay == 0 {
		c.Batch.RetryDelay = time.Second * 2
	}
}

// Validate は拡張設定を検証する
func (c *EnhancedConfig) Validate() error {
	// 基本設定の検証
	if err := c.Config.Validate(); err != nil {
		return err
	}

	// 出力フォーマットの検証
	validFormats := map[string]bool{"json": true, "yaml": true, "table": true}
	if !validFormats[c.OutputFormat] {
		return fmt.Errorf("無効な出力フォーマット: %s (有効な値: json, yaml, table)", c.OutputFormat)
	}

	// ログレベルの検証
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("無効なログレベル: %s (有効な値: debug, info, warn, error)", c.Logging.Level)
	}

	// ログフォーマットの検証
	validLogFormats := map[string]bool{"json": true, "text": true}
	if !validLogFormats[c.Logging.Format] {
		return fmt.Errorf("無効なログフォーマット: %s (有効な値: json, text)", c.Logging.Format)
	}

	// バッチ設定の検証
	if c.Batch.MaxConcurrency < 1 {
		return fmt.Errorf("同時実行数は1以上である必要があります")
	}
	if c.Batch.RetryAttempts < 0 {
		return fmt.Errorf("リトライ回数は0以上である必要があります")
	}

	return nil
}

// MergeWithEnvironment は環境変数で設定を上書きする
func (c *EnhancedConfig) MergeWithEnvironment() {
	if region := os.Getenv("PHANTOM_ECS_REGION"); region != "" {
		c.Region = region
	}
	if profile := os.Getenv("PHANTOM_ECS_PROFILE"); profile != "" {
		c.Profile = profile
	}
	if format := os.Getenv("PHANTOM_ECS_OUTPUT_FORMAT"); format != "" {
		c.OutputFormat = format
	}
	if level := os.Getenv("PHANTOM_ECS_LOG_LEVEL"); level != "" {
		c.Logging.Level = level
	}
	if format := os.Getenv("PHANTOM_ECS_LOG_FORMAT"); format != "" {
		c.Logging.Format = format
	}
	if filename := os.Getenv("PHANTOM_ECS_LOG_FILE"); filename != "" {
		c.Logging.Filename = filename
	}
	if maxSize := getEnvInt("PHANTOM_ECS_LOG_MAX_SIZE"); maxSize > 0 {
		c.Logging.MaxSize = maxSize
	}
	if maxAge := getEnvInt("PHANTOM_ECS_LOG_MAX_AGE"); maxAge > 0 {
		c.Logging.MaxAge = maxAge
	}
	if maxBackups := getEnvInt("PHANTOM_ECS_LOG_MAX_BACKUPS"); maxBackups > 0 {
		c.Logging.MaxBackups = maxBackups
	}
	if maxConcurrency := getEnvInt("PHANTOM_ECS_BATCH_MAX_CONCURRENCY"); maxConcurrency > 0 {
		c.Batch.MaxConcurrency = maxConcurrency
	}
	if retryAttempts := getEnvInt("PHANTOM_ECS_BATCH_RETRY_ATTEMPTS"); retryAttempts >= 0 {
		c.Batch.RetryAttempts = retryAttempts
	}
	if retryDelay := getEnvDuration("PHANTOM_ECS_BATCH_RETRY_DELAY"); retryDelay > 0 {
		c.Batch.RetryDelay = retryDelay
	}
	if showProgress := getEnvBool("PHANTOM_ECS_BATCH_SHOW_PROGRESS"); showProgress != nil {
		c.Batch.ShowProgress = *showProgress
	}
}

// SaveToFile は設定をYAMLファイルに保存する
func (c *EnhancedConfig) SaveToFile(filename string) error {
	fileConfig := FileConfig{
		Profiles: map[string]ProfileConfig{
			"default": {
				Region:       c.Region,
				OutputFormat: c.OutputFormat,
				AWSProfile:   c.Profile,
			},
		},
		Logging: c.Logging,
		Batch:   c.Batch,
	}

	data, err := yaml.Marshal(&fileConfig)
	if err != nil {
		return fmt.Errorf("YAML生成に失敗しました: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("ファイル書き込みに失敗しました: %w", err)
	}

	return nil
}

// ヘルパー関数

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := getEnvInt(key); value != 0 {
		return value
	}
	return defaultValue
}

func getEnvInt(key string) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return 0
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := getEnvBool(key); value != nil {
		return *value
	}
	return defaultValue
}

func getEnvBool(key string) *bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return &boolValue
		}
	}
	return nil
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := getEnvDuration(key); value != 0 {
		return value
	}
	return defaultValue
}

func getEnvDuration(key string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return 0
}
