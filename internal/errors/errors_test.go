package errors_test

import (
	"errors"
	"testing"

	phantomecs_errors "github.com/dev-shimada/phantom-ecs/internal/errors"

	"github.com/stretchr/testify/assert"
)

func TestNewPhantomError(t *testing.T) {
	tests := []struct {
		name     string
		errType  phantomecs_errors.ErrorType
		message  string
		cause    error
		expected string
	}{
		{
			name:     "基本的なエラー作成",
			errType:  phantomecs_errors.ErrTypeConfig,
			message:  "設定ファイルが見つかりません",
			cause:    nil,
			expected: "設定ファイルが見つかりません",
		},
		{
			name:     "原因付きエラー作成",
			errType:  phantomecs_errors.ErrTypeAWS,
			message:  "AWS API呼び出しに失敗しました",
			cause:    errors.New("network timeout"),
			expected: "AWS API呼び出しに失敗しました: network timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := phantomecs_errors.NewPhantomError(tt.errType, tt.message, tt.cause)

			assert.Equal(t, tt.errType, err.Type)
			assert.Equal(t, tt.expected, err.Error())
			assert.Equal(t, tt.cause, err.Cause)
		})
	}
}

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		errType  phantomecs_errors.ErrorType
		expected int
	}{
		{
			name:     "設定エラーの終了コード",
			errType:  phantomecs_errors.ErrTypeConfig,
			expected: 1,
		},
		{
			name:     "AWSエラーの終了コード",
			errType:  phantomecs_errors.ErrTypeAWS,
			expected: 2,
		},
		{
			name:     "バリデーションエラーの終了コード",
			errType:  phantomecs_errors.ErrTypeValidation,
			expected: 3,
		},
		{
			name:     "ネットワークエラーの終了コード",
			errType:  phantomecs_errors.ErrTypeNetwork,
			expected: 4,
		},
		{
			name:     "一般エラーの終了コード",
			errType:  phantomecs_errors.ErrTypeGeneral,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := phantomecs_errors.NewPhantomError(tt.errType, "test error", nil)
			assert.Equal(t, tt.expected, err.GetExitCode())
		})
	}
}

func TestIsPhantomError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "PhantomErrorの場合",
			err:      phantomecs_errors.NewPhantomError(phantomecs_errors.ErrTypeConfig, "test", nil),
			expected: true,
		},
		{
			name:     "標準エラーの場合",
			err:      errors.New("standard error"),
			expected: false,
		},
		{
			name:     "nilの場合",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := phantomecs_errors.IsPhantomError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := phantomecs_errors.WrapError(phantomecs_errors.ErrTypeAWS, "操作に失敗しました", originalErr)

	assert.Equal(t, phantomecs_errors.ErrTypeAWS, wrappedErr.Type)
	assert.Equal(t, "操作に失敗しました: original error", wrappedErr.Error())
	assert.Equal(t, originalErr, wrappedErr.Cause)
}

func TestErrorChaining(t *testing.T) {
	cause := errors.New("root cause")
	err1 := phantomecs_errors.WrapError(phantomecs_errors.ErrTypeNetwork, "network error", cause)
	err2 := phantomecs_errors.WrapError(phantomecs_errors.ErrTypeAWS, "aws error", err1)

	assert.Contains(t, err2.Error(), "aws error")
	assert.Contains(t, err2.Error(), "network error")
	assert.Contains(t, err2.Error(), "root cause")
}
