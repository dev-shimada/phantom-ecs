package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/schollz/progressbar/v3"
)

// Config はバッチ処理の設定
type Config struct {
	// MaxConcurrency は同時実行数の上限
	MaxConcurrency int
	// RetryAttempts はリトライ回数
	RetryAttempts int
	// RetryDelay はリトライ間隔
	RetryDelay time.Duration
	// ShowProgress はプログレスバーの表示フラグ
	ShowProgress bool
}

// Processor はバッチ処理で実行される処理のインターフェース
type Processor interface {
	Process(ctx context.Context, service string) error
}

// ProcessResult はバッチ処理の結果
type ProcessResult struct {
	ServiceName string
	Success     bool
	Error       error
	Duration    time.Duration
}

// BatchProcessor はバッチ処理を管理する
type BatchProcessor struct {
	config    *Config
	processor Processor
}

// NewBatchProcessor は新しいバッチプロセッサを作成する
func NewBatchProcessor(config *Config, processor Processor) *BatchProcessor {
	return &BatchProcessor{
		config:    config,
		processor: processor,
	}
}

// ProcessServices は複数のサービスを並列処理する
func (bp *BatchProcessor) ProcessServices(ctx context.Context, services []string) ([]*ProcessResult, error) {
	results := make([]*ProcessResult, len(services))

	// プログレスバーの設定
	var bar *progressbar.ProgressBar
	if bp.config.ShowProgress {
		bar = progressbar.NewOptions(len(services),
			progressbar.OptionSetDescription("Processing services..."),
			progressbar.OptionSetWidth(15),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerHead:    ">",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)
	}

	// セマフォで同時実行数を制限
	semaphore := make(chan struct{}, bp.config.MaxConcurrency)
	var wg sync.WaitGroup

	for i, service := range services {
		wg.Add(1)
		go func(index int, serviceName string) {
			defer wg.Done()

			// セマフォで同時実行数を制限
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := bp.processServiceWithRetry(ctx, serviceName)
			results[index] = result

			// プログレスバーの更新
			if bar != nil {
				bar.Add(1)
			}
		}(i, service)
	}

	wg.Wait()

	if bar != nil {
		bar.Finish()
	}

	return results, nil
}

// processServiceWithRetry はリトライ機能付きでサービスを処理する
func (bp *BatchProcessor) processServiceWithRetry(ctx context.Context, serviceName string) *ProcessResult {
	start := time.Now()

	var lastErr error
	err := retry.Do(
		func() error {
			err := bp.processor.Process(ctx, serviceName)
			if err != nil {
				lastErr = err
				return err
			}
			return nil
		},
		retry.Attempts(uint(bp.config.RetryAttempts+1)), // 初回 + リトライ回数
		retry.Delay(bp.config.RetryDelay),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) {
			// リトライ時のログ（必要に応じて）
		}),
	)

	duration := time.Since(start)

	if err != nil {
		return &ProcessResult{
			ServiceName: serviceName,
			Success:     false,
			Error:       lastErr,
			Duration:    duration,
		}
	}

	return &ProcessResult{
		ServiceName: serviceName,
		Success:     true,
		Error:       nil,
		Duration:    duration,
	}
}

// GetDefaultConfig はデフォルト設定を返す
func GetDefaultConfig() *Config {
	return &Config{
		MaxConcurrency: 3,
		RetryAttempts:  3,
		RetryDelay:     time.Second * 2,
		ShowProgress:   true,
	}
}

// ProcessorFunc は関数をProcessorインターフェースに変換する
type ProcessorFunc func(ctx context.Context, service string) error

// Process はProcessorインターフェースの実装
func (f ProcessorFunc) Process(ctx context.Context, service string) error {
	return f(ctx, service)
}

// Statistics はバッチ処理の統計情報
type Statistics struct {
	TotalServices   int
	SuccessfulCount int
	FailedCount     int
	TotalDuration   time.Duration
	AverageDuration time.Duration
	FailedServices  []string
}

// CalculateStatistics は処理結果から統計情報を計算する
func CalculateStatistics(results []*ProcessResult) *Statistics {
	stats := &Statistics{
		TotalServices:  len(results),
		FailedServices: make([]string, 0),
	}

	var totalDuration time.Duration
	for _, result := range results {
		totalDuration += result.Duration

		if result.Success {
			stats.SuccessfulCount++
		} else {
			stats.FailedCount++
			stats.FailedServices = append(stats.FailedServices, result.ServiceName)
		}
	}

	stats.TotalDuration = totalDuration
	if len(results) > 0 {
		stats.AverageDuration = totalDuration / time.Duration(len(results))
	}

	return stats
}

// PrintStatistics は統計情報を表示する
func (s *Statistics) PrintStatistics() {
	fmt.Printf("\n=== バッチ処理統計 ===\n")
	fmt.Printf("総サービス数: %d\n", s.TotalServices)
	fmt.Printf("成功: %d\n", s.SuccessfulCount)
	fmt.Printf("失敗: %d\n", s.FailedCount)
	fmt.Printf("総処理時間: %v\n", s.TotalDuration)
	fmt.Printf("平均処理時間: %v\n", s.AverageDuration)

	if len(s.FailedServices) > 0 {
		fmt.Printf("\n失敗したサービス:\n")
		for _, service := range s.FailedServices {
			fmt.Printf("  - %s\n", service)
		}
	}
	fmt.Printf("====================\n")
}
