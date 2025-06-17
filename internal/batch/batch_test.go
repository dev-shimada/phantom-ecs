package batch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProcessor はテスト用のモックプロセッサ
type MockProcessor struct {
	mock.Mock
}

func (m *MockProcessor) Process(ctx context.Context, service string) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func TestNewBatchProcessor(t *testing.T) {
	config := &Config{
		MaxConcurrency: 5,
		RetryAttempts:  3,
		RetryDelay:     time.Second,
	}

	processor := &MockProcessor{}
	batchProcessor := NewBatchProcessor(config, processor)

	assert.NotNil(t, batchProcessor)
	assert.Equal(t, config, batchProcessor.config)
	assert.Equal(t, processor, batchProcessor.processor)
}

func TestProcessServices_Success(t *testing.T) {
	config := &Config{
		MaxConcurrency: 2,
		RetryAttempts:  1,
		RetryDelay:     time.Millisecond * 10,
	}

	processor := &MockProcessor{}
	services := []string{"service1", "service2", "service3"}

	// 全てのサービスで成功を期待
	for _, service := range services {
		processor.On("Process", mock.Anything, service).Return(nil)
	}

	batchProcessor := NewBatchProcessor(config, processor)
	ctx := context.Background()

	results, err := batchProcessor.ProcessServices(ctx, services)

	require.NoError(t, err)
	assert.Len(t, results, 3)

	for _, result := range results {
		assert.NoError(t, result.Error)
		assert.True(t, result.Success)
	}

	processor.AssertExpectations(t)
}

func TestProcessServices_WithErrors(t *testing.T) {
	config := &Config{
		MaxConcurrency: 2,
		RetryAttempts:  1,
		RetryDelay:     time.Millisecond * 10,
	}

	processor := &MockProcessor{}
	services := []string{"service1", "service2", "service3"}

	// service1は成功、service2は失敗、service3は成功
	processor.On("Process", mock.Anything, "service1").Return(nil)
	processor.On("Process", mock.Anything, "service2").Return(errors.New("処理失敗"))
	processor.On("Process", mock.Anything, "service3").Return(nil)

	batchProcessor := NewBatchProcessor(config, processor)
	ctx := context.Background()

	results, err := batchProcessor.ProcessServices(ctx, services)

	require.NoError(t, err)
	assert.Len(t, results, 3)

	// service1は成功
	assert.True(t, results[0].Success)
	assert.NoError(t, results[0].Error)

	// service2は失敗
	assert.False(t, results[1].Success)
	assert.Error(t, results[1].Error)
	assert.Contains(t, results[1].Error.Error(), "処理失敗")

	// service3は成功
	assert.True(t, results[2].Success)
	assert.NoError(t, results[2].Error)

	processor.AssertExpectations(t)
}

func TestProcessServices_WithRetry(t *testing.T) {
	config := &Config{
		MaxConcurrency: 1,
		RetryAttempts:  3,
		RetryDelay:     time.Millisecond * 10,
	}

	processor := &MockProcessor{}
	services := []string{"service1"}

	// 最初の2回は失敗、3回目で成功
	processor.On("Process", mock.Anything, "service1").Return(errors.New("一時的な失敗")).Times(2)
	processor.On("Process", mock.Anything, "service1").Return(nil).Once()

	batchProcessor := NewBatchProcessor(config, processor)
	ctx := context.Background()

	results, err := batchProcessor.ProcessServices(ctx, services)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.True(t, results[0].Success)
	assert.NoError(t, results[0].Error)

	processor.AssertExpectations(t)
}

func TestProcessServices_ExceedsRetryLimit(t *testing.T) {
	config := &Config{
		MaxConcurrency: 1,
		RetryAttempts:  2,
		RetryDelay:     time.Millisecond * 10,
	}

	processor := &MockProcessor{}
	services := []string{"service1"}

	// 全ての試行で失敗
	processor.On("Process", mock.Anything, "service1").Return(errors.New("恒久的な失敗")).Times(3)

	batchProcessor := NewBatchProcessor(config, processor)
	ctx := context.Background()

	results, err := batchProcessor.ProcessServices(ctx, services)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.Error(t, results[0].Error)
	assert.Contains(t, results[0].Error.Error(), "恒久的な失敗")

	processor.AssertExpectations(t)
}

func TestProcessServices_ContextCancellation(t *testing.T) {
	config := &Config{
		MaxConcurrency: 1,
		RetryAttempts:  1,
		RetryDelay:     time.Second, // 長いディレイ
	}

	processor := &MockProcessor{}
	services := []string{"service1"}

	// キャンセレーションをシミュレート
	processor.On("Process", mock.Anything, "service1").Return(context.Canceled)

	batchProcessor := NewBatchProcessor(config, processor)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	results, err := batchProcessor.ProcessServices(ctx, services)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.Error(t, results[0].Error)

	processor.AssertExpectations(t)
}

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	assert.Equal(t, 3, config.MaxConcurrency)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, time.Second*2, config.RetryDelay)
	assert.True(t, config.ShowProgress)
}

func TestProcessResult(t *testing.T) {
	// 成功ケース
	successResult := &ProcessResult{
		ServiceName: "test-service",
		Success:     true,
		Error:       nil,
		Duration:    time.Second,
	}

	assert.Equal(t, "test-service", successResult.ServiceName)
	assert.True(t, successResult.Success)
	assert.NoError(t, successResult.Error)
	assert.Equal(t, time.Second, successResult.Duration)

	// 失敗ケース
	failureResult := &ProcessResult{
		ServiceName: "test-service",
		Success:     false,
		Error:       errors.New("テストエラー"),
		Duration:    time.Millisecond * 500,
	}

	assert.Equal(t, "test-service", failureResult.ServiceName)
	assert.False(t, failureResult.Success)
	assert.Error(t, failureResult.Error)
	assert.Equal(t, time.Millisecond*500, failureResult.Duration)
}
