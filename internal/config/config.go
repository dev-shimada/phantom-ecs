package config

import (
	"fmt"
	"os"
	"strings"
)

// Config アプリケーション設定を表す構造体
type Config struct {
	Region       string
	Profile      string
	OutputFormat string
}

// validRegions 有効なAWSリージョンのリスト (設計仕様書には記載がなかったため、一般的なものをいくつか追加)
var validRegions = map[string]struct{}{
	"us-east-1":      {},
	"us-east-2":      {},
	"us-west-1":      {},
	"us-west-2":      {},
	"af-south-1":     {},
	"ap-east-1":      {},
	"ap-south-1":     {},
	"ap-northeast-3": {},
	"ap-northeast-2": {},
	"ap-southeast-1": {},
	"ap-southeast-2": {},
	"ap-northeast-1": {},
	"ca-central-1":   {},
	"eu-central-1":   {},
	"eu-west-1":      {},
	"eu-west-2":      {},
	"eu-south-1":     {},
	"eu-west-3":      {},
	"eu-north-1":     {},
	"me-south-1":     {},
	"sa-east-1":      {},
}

const (
	DefaultRegion       = "us-east-1"
	DefaultOutputFormat = "table"
)

// NewConfig 新しい設定オブジェクトを作成
func NewConfig(region, profile string) *Config {
	if region == "" {
		region = DefaultRegion
	}
	return &Config{
		Region:       region,
		Profile:      profile,
		OutputFormat: DefaultOutputFormat,
	}
}

// NewConfigFromEnvironment 環境変数から設定オブジェクトを作成
func NewConfigFromEnvironment() *Config {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = DefaultRegion
	}
	profile := os.Getenv("AWS_PROFILE")
	return &Config{
		Region:       region,
		Profile:      profile,
		OutputFormat: DefaultOutputFormat,
	}
}

// GetRegion リージョンを取得
func (c *Config) GetRegion() string {
	return c.Region
}

// GetProfile プロファイルを取得
func (c *Config) GetProfile() string {
	return c.Profile
}

// GetOutputFormat 出力フォーマットを取得
func (c *Config) GetOutputFormat() string {
	return c.OutputFormat
}

// SetOutputFormat 出力フォーマットを設定
func (c *Config) SetOutputFormat(format string) {
	lowerFormat := strings.ToLower(format)
	if lowerFormat == "json" || lowerFormat == "yaml" || lowerFormat == "table" {
		c.OutputFormat = lowerFormat
	}
	// 不正な値の場合はデフォルト(table)のまま
}

// Validate 設定を検証
func (c *Config) Validate() error {
	if _, ok := validRegions[c.Region]; !ok {
		return fmt.Errorf("invalid AWS region: %s", c.Region)
	}
	return nil
}
