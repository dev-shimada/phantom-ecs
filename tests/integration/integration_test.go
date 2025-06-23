package integration_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dev-shimada/phantom-ecs/internal/batch"
	"github.com/dev-shimada/phantom-ecs/internal/config"
	"github.com/dev-shimada/phantom-ecs/internal/errors"
	"github.com/dev-shimada/phantom-ecs/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockECSService はテスト用のECSサービスモック
type MockECSService struct {
	Services []string
	Failures map[string]error
}

func (m *MockECSService) Process(ctx context.Context, service string) error {
	if err, exists := m.Failures[service]; exists {
		return err
	}
	// サービス処理のシミュレーション
	time.Sleep(time.Millisecond * 100)
	return nil
}

func TestEndToEndBatchProcessing(t *testing.T) {
	// テスト用設定
	config := &batch.Config{
		MaxConcurrency: 2,
		RetryAttempts:  1,
		RetryDelay:     time.Millisecond * 50,
		ShowProgress:   false, // テストではプログレスバーを非表示
	}

	// モックサービス
	mockService := &MockECSService{
		Services: []string{"service1", "service2", "service3", "service4"},
		Failures: map[string]error{
			"service2": errors.NewAWSError("一時的な失敗", nil),
		},
	}

	// バッチプロセッサ
	processor := batch.NewBatchProcessor(config, mockService)

	// 実行
	ctx := context.Background()
	results, err := processor.ProcessServices(ctx, mockService.Services)

	require.NoError(t, err)
	assert.Len(t, results, 4)

	// 結果の検証
	successCount := 0
	failureCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	assert.Equal(t, 3, successCount) // service1, service3, service4が成功
	assert.Equal(t, 1, failureCount) // service2が失敗

	// 統計の確認
	stats := batch.CalculateStatistics(results)
	assert.Equal(t, 4, stats.TotalServices)
	assert.Equal(t, 3, stats.SuccessfulCount)
	assert.Equal(t, 1, stats.FailedCount)
	assert.Contains(t, stats.FailedServices, "service2")
}

func TestConfigurationIntegration(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	// YAML設定ファイルを作成
	yamlContent := `
profiles:
  test:
    region: ap-northeast-1
    output_format: json
    aws_profile: test-profile

logging:
  level: debug
  format: json
  max_size: 50

batch:
  max_concurrency: 4
  retry_attempts: 2
  retry_delay: 1s
  show_progress: false
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// 設定を読み込み
	enhancedConfig, err := config.LoadFromFile(configFile, "test")
	require.NoError(t, err)

	// 設定値の検証
	assert.Equal(t, "ap-northeast-1", enhancedConfig.Region)
	assert.Equal(t, "json", enhancedConfig.OutputFormat)
	assert.Equal(t, "test-profile", enhancedConfig.Profile)
	assert.Equal(t, "debug", enhancedConfig.Logging.Level)
	assert.Equal(t, "json", enhancedConfig.Logging.Format)
	assert.Equal(t, 50, enhancedConfig.Logging.MaxSize)
	assert.Equal(t, 4, enhancedConfig.Batch.MaxConcurrency)
	assert.Equal(t, 2, enhancedConfig.Batch.RetryAttempts)
	assert.Equal(t, time.Second, enhancedConfig.Batch.RetryDelay)
	assert.False(t, enhancedConfig.Batch.ShowProgress)

	// バリデーション
	err = enhancedConfig.Validate()
	assert.NoError(t, err)
}

func TestLoggingIntegration(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "integration-test.log")

	// ロガー設定
	logConfig := &logger.Config{
		Level:      "debug",
		Format:     "json",
		Filename:   logFile,
		MaxSize:    1,
		MaxAge:     1,
		MaxBackups: 1,
	}

	// ロガー作成
	log, err := logger.NewLogger(logConfig)
	require.NoError(t, err)

	// ログ出力テスト
	log.Info("統合テスト開始")
	log.WithFields(map[string]interface{}{
		"service": "test-service",
		"region":  "us-east-1",
	}).Debug("サービス処理中")
	log.Error("テストエラー")

	// ファイルが作成されたことを確認
	assert.FileExists(t, logFile)

	// ファイル内容の確認（簡単なチェック）
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "統合テスト開始")
	assert.Contains(t, contentStr, "test-service")
	assert.Contains(t, contentStr, "テストエラー")
}

func TestErrorHandlingIntegration(t *testing.T) {
	// カスタムエラーの作成
	configErr := errors.NewConfigError("設定ファイルが見つかりません", nil)
	awsErr := errors.NewAWSError("AWS API呼び出しエラー", fmt.Errorf("network timeout"))
	validationErr := errors.NewValidationError("無効なパラメータ", nil)

	// エラータイプの確認
	assert.True(t, errors.IsPhantomError(configErr))
	assert.True(t, errors.IsPhantomError(awsErr))
	assert.True(t, errors.IsPhantomError(validationErr))

	// 終了コードの確認
	assert.Equal(t, 1, configErr.GetExitCode())
	assert.Equal(t, 2, awsErr.GetExitCode())
	assert.Equal(t, 3, validationErr.GetExitCode())

	// エラーメッセージの確認
	assert.Equal(t, "設定ファイルが見つかりません", configErr.Error())
	assert.Contains(t, awsErr.Error(), "AWS API呼び出しエラー")
	assert.Contains(t, awsErr.Error(), "network timeout")
}

func TestFullWorkflow(t *testing.T) {
	// ステップ1: 設定の読み込み
	enhancedConfig := config.GetDefaultEnhancedConfig()
	enhancedConfig.Region = "us-west-2"
	enhancedConfig.Batch.MaxConcurrency = 2
	enhancedConfig.Batch.ShowProgress = false
	enhancedConfig.Logging.Level = "info"

	err := enhancedConfig.Validate()
	require.NoError(t, err)

	// ステップ2: ロガーの初期化
	tempDir := t.TempDir()
	enhancedConfig.Logging.Filename = filepath.Join(tempDir, "workflow.log")

	log, err := logger.NewLogger(&logger.Config{
		Level:      enhancedConfig.Logging.Level,
		Format:     enhancedConfig.Logging.Format,
		Filename:   enhancedConfig.Logging.Filename,
		MaxSize:    enhancedConfig.Logging.MaxSize,
		MaxAge:     enhancedConfig.Logging.MaxAge,
		MaxBackups: enhancedConfig.Logging.MaxBackups,
	})
	require.NoError(t, err)

	log.Info("ワークフロー開始")

	// ステップ3: バッチ処理の実行
	batchConfig := &batch.Config{
		MaxConcurrency: enhancedConfig.Batch.MaxConcurrency,
		RetryAttempts:  enhancedConfig.Batch.RetryAttempts,
		RetryDelay:     enhancedConfig.Batch.RetryDelay,
		ShowProgress:   enhancedConfig.Batch.ShowProgress,
	}

	mockService := &MockECSService{
		Services: []string{"web-service", "api-service", "worker-service"},
	}

	processor := batch.NewBatchProcessor(batchConfig, mockService)
	ctx := context.Background()

	log.Info("バッチ処理開始")
	results, err := processor.ProcessServices(ctx, mockService.Services)
	require.NoError(t, err)

	// ステップ4: 結果の検証とログ出力
	stats := batch.CalculateStatistics(results)
	log.WithFields(map[string]interface{}{
		"total_services": stats.TotalServices,
		"successful":     stats.SuccessfulCount,
		"failed":         stats.FailedCount,
		"duration":       stats.TotalDuration.String(),
	}).Info("バッチ処理完了")

	// 検証
	assert.Equal(t, 3, stats.TotalServices)
	assert.Equal(t, 3, stats.SuccessfulCount)
	assert.Equal(t, 0, stats.FailedCount)

	// ログファイルの確認
	assert.FileExists(t, enhancedConfig.Logging.Filename)
}

func TestPerformance(t *testing.T) {
	// パフォーマンステスト（大量のサービス処理）
	const serviceCount = 100

	config := &batch.Config{
		MaxConcurrency: 10,
		RetryAttempts:  1,
		RetryDelay:     time.Millisecond * 10,
		ShowProgress:   false,
	}

	services := make([]string, serviceCount)
	for i := 0; i < serviceCount; i++ {
		services[i] = fmt.Sprintf("service-%d", i)
	}

	mockService := &MockECSService{
		Services: services,
		// いくつかのサービスで失敗をシミュレート
		Failures: map[string]error{
			"service-10": errors.NewNetworkError("一時的なネットワークエラー", nil),
			"service-25": errors.NewAWSError("レート制限", nil),
			"service-50": errors.NewNetworkError("タイムアウト", nil),
		},
	}

	processor := batch.NewBatchProcessor(config, mockService)

	start := time.Now()
	ctx := context.Background()
	results, err := processor.ProcessServices(ctx, services)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, results, serviceCount)

	stats := batch.CalculateStatistics(results)
	assert.Equal(t, serviceCount, stats.TotalServices)
	assert.Equal(t, serviceCount-3, stats.SuccessfulCount) // 3つが失敗
	assert.Equal(t, 3, stats.FailedCount)

	// パフォーマンス確認（同時実行により短時間で完了すること）
	// 単純な逐次処理だと100 * 100ms = 10秒かかるが、
	// 10並列実行により大幅に短縮されることを確認
	maxExpectedDuration := time.Second * 2 // 余裕を持った上限
	assert.Less(t, duration, maxExpectedDuration,
		"パフォーマンステスト失敗: 処理時間が%v、期待値は%v未満", duration, maxExpectedDuration)

	t.Logf("パフォーマンステスト結果: %d サービスを %v で処理完了", serviceCount, duration)
}
