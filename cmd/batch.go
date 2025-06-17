package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dev-shimada/phantom-ecs/internal/batch"
	"github.com/dev-shimada/phantom-ecs/internal/config"
	"github.com/dev-shimada/phantom-ecs/internal/errors"
	"github.com/dev-shimada/phantom-ecs/internal/logger"
	"github.com/spf13/cobra"
)

var (
	batchConfigFile   string
	batchProfile      string
	batchServices     []string
	batchConcurrency  int
	batchRetryCount   int
	batchRetryDelay   time.Duration
	batchShowProgress bool
	batchDryRun       bool
)

// NewBatchCommand はバッチ処理コマンドを作成する
func NewBatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "複数のECSサービスをバッチ処理します",
		Long: `複数のECSサービスに対して指定された操作をバッチ処理で実行します。
設定ファイルからサービスリストを読み込むか、コマンドラインで直接指定できます。

例:
  phantom-ecs batch --services service1,service2,service3
  phantom-ecs batch --config-file batch-config.yaml --profile production
  phantom-ecs batch --services service1,service2 --concurrency 5 --retry-count 3`,
		RunE: runBatch,
	}

	cmd.Flags().StringVar(&batchConfigFile, "config-file", "", "バッチ設定ファイルのパス")
	cmd.Flags().StringVar(&batchProfile, "batch-profile", "default", "使用するバッチプロファイル")
	cmd.Flags().StringSliceVar(&batchServices, "services", []string{}, "処理対象のサービス名（カンマ区切り）")
	cmd.Flags().IntVar(&batchConcurrency, "concurrency", 3, "同時実行数")
	cmd.Flags().IntVar(&batchRetryCount, "retry-count", 3, "リトライ回数")
	cmd.Flags().DurationVar(&batchRetryDelay, "retry-delay", time.Second*2, "リトライ間隔")
	cmd.Flags().BoolVar(&batchShowProgress, "progress", true, "プログレスバーを表示")
	cmd.Flags().BoolVar(&batchDryRun, "dry-run", false, "実際には実行せず、処理内容のみ表示")

	return cmd
}

func runBatch(cmd *cobra.Command, args []string) error {
	// ロガーの初期化
	log, err := logger.NewDefaultLogger()
	if err != nil {
		return errors.NewGeneralError("ロガーの初期化に失敗しました", err)
	}

	log.Info("バッチ処理を開始します")

	// 設定の読み込み
	var enhancedConfig *config.EnhancedConfig
	if batchConfigFile != "" {
		enhancedConfig, err = config.LoadFromFile(batchConfigFile, batchProfile)
		if err != nil {
			return errors.NewConfigError("設定ファイルの読み込みに失敗しました", err)
		}
	} else {
		enhancedConfig = config.GetDefaultEnhancedConfig()
	}

	// コマンドライン引数で設定を上書き
	if cmd.Flags().Changed("concurrency") {
		enhancedConfig.Batch.MaxConcurrency = batchConcurrency
	}
	if cmd.Flags().Changed("retry-count") {
		enhancedConfig.Batch.RetryAttempts = batchRetryCount
	}
	if cmd.Flags().Changed("retry-delay") {
		enhancedConfig.Batch.RetryDelay = batchRetryDelay
	}
	if cmd.Flags().Changed("progress") {
		enhancedConfig.Batch.ShowProgress = batchShowProgress
	}

	// 環境変数での上書き
	enhancedConfig.MergeWithEnvironment()

	// 設定の検証
	if err := enhancedConfig.Validate(); err != nil {
		return errors.NewValidationError("設定の検証に失敗しました", err)
	}

	// サービスリストの取得
	var services []string
	if len(batchServices) > 0 {
		services = batchServices
	} else {
		// サービスリストが指定されていない場合はエラー
		return errors.NewValidationError("処理対象のサービスを指定してください（--servicesフラグ）", nil)
	}

	if len(services) == 0 {
		return errors.NewValidationError("処理対象のサービスが見つかりません", nil)
	}

	log.WithFields(map[string]interface{}{
		"service_count": len(services),
		"services":      strings.Join(services, ", "),
	}).Info("バッチ処理対象サービス")

	// Dry runモードの場合
	if batchDryRun {
		fmt.Printf("=== Dry Run モード ===\n")
		fmt.Printf("処理対象サービス数: %d\n", len(services))
		fmt.Printf("同時実行数: %d\n", enhancedConfig.Batch.MaxConcurrency)
		fmt.Printf("リトライ回数: %d\n", enhancedConfig.Batch.RetryAttempts)
		fmt.Printf("リトライ間隔: %v\n", enhancedConfig.Batch.RetryDelay)
		fmt.Printf("\n処理対象サービス:\n")
		for i, service := range services {
			fmt.Printf("  %d. %s\n", i+1, service)
		}
		fmt.Printf("\n実際の処理は実行されません。\n")
		return nil
	}

	// バッチ処理の実行
	processor := &BatchServiceProcessor{
		config: enhancedConfig,
		logger: log,
	}

	batchConfig := &batch.Config{
		MaxConcurrency: enhancedConfig.Batch.MaxConcurrency,
		RetryAttempts:  enhancedConfig.Batch.RetryAttempts,
		RetryDelay:     enhancedConfig.Batch.RetryDelay,
		ShowProgress:   enhancedConfig.Batch.ShowProgress,
	}

	batchProcessor := batch.NewBatchProcessor(batchConfig, processor)

	ctx := context.Background()
	start := time.Now()

	results, err := batchProcessor.ProcessServices(ctx, services)
	if err != nil {
		return errors.NewGeneralError("バッチ処理に失敗しました", err)
	}

	duration := time.Since(start)

	// 結果の表示
	stats := batch.CalculateStatistics(results)

	fmt.Printf("\n=== バッチ処理結果 ===\n")
	fmt.Printf("総処理時間: %v\n", duration)
	fmt.Printf("総サービス数: %d\n", stats.TotalServices)
	fmt.Printf("成功: %d\n", stats.SuccessfulCount)
	fmt.Printf("失敗: %d\n", stats.FailedCount)
	fmt.Printf("平均処理時間: %v\n", stats.AverageDuration)

	if len(stats.FailedServices) > 0 {
		fmt.Printf("\n失敗したサービス:\n")
		for _, service := range stats.FailedServices {
			fmt.Printf("  - %s\n", service)
		}
	}

	// ログ出力
	log.WithFields(map[string]interface{}{
		"total_duration":   duration.String(),
		"total_services":   stats.TotalServices,
		"successful_count": stats.SuccessfulCount,
		"failed_count":     stats.FailedCount,
		"average_duration": stats.AverageDuration.String(),
	}).Info("バッチ処理が完了しました")

	// 失敗があった場合は非ゼロ終了コード
	if stats.FailedCount > 0 {
		os.Exit(1)
	}

	return nil
}

// BatchServiceProcessor はバッチ処理用のサービスプロセッサ
type BatchServiceProcessor struct {
	config *config.EnhancedConfig
	logger logger.Logger
}

// Process はサービスを処理する（現在は基本的な検査のみ）
func (p *BatchServiceProcessor) Process(ctx context.Context, serviceName string) error {
	p.logger.WithFields(map[string]interface{}{
		"service": serviceName,
		"region":  p.config.Region,
	}).Info("サービス処理開始")

	// ここでは実際のサービス処理をシミュレート
	// 実際の実装では、inspectやdeployなどの処理を実行
	time.Sleep(time.Millisecond * 100) // 処理時間のシミュレート

	p.logger.WithFields(map[string]interface{}{
		"service": serviceName,
		"region":  p.config.Region,
	}).Info("サービス処理完了")

	return nil
}
