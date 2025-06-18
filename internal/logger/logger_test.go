package logger_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dev-shimada/phantom-ecs/internal/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name     string
		config   logger.Config
		expected logrus.Level
	}{
		{
			name: "デバッグレベル設定",
			config: logger.Config{
				Level:  "debug",
				Format: "json",
			},
			expected: logrus.DebugLevel,
		},
		{
			name: "インフォレベル設定",
			config: logger.Config{
				Level:  "info",
				Format: "text",
			},
			expected: logrus.InfoLevel,
		},
		{
			name: "エラーレベル設定",
			config: logger.Config{
				Level:  "error",
				Format: "json",
			},
			expected: logrus.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := logger.NewLogger(&tt.config)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, logger.GetLevel())
		})
	}
}

func TestLoggerWithJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  "info",
		Format: "json",
		Output: &buf,
	}

	logger, err := logger.NewLogger(config)
	require.NoError(t, err)

	logger.Info("テストメッセージ")

	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "info", logEntry["level"])
	assert.Equal(t, "テストメッセージ", logEntry["msg"])
	assert.Contains(t, logEntry, "time")
}

func TestLoggerWithTextFormat(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  "info",
		Format: "text",
		Output: &buf,
	}

	logger, err := logger.NewLogger(config)
	require.NoError(t, err)

	logger.Info("テストメッセージ")

	output := buf.String()
	assert.Contains(t, output, "level=info")
	assert.Contains(t, output, "テストメッセージ")
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  "info",
		Format: "json",
		Output: &buf,
	}

	logger, err := logger.NewLogger(config)
	require.NoError(t, err)

	logger.WithFields(logrus.Fields{
		"service": "test-service",
		"cluster": "test-cluster",
	}).Info("サービス情報")

	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "test-service", logEntry["service"])
	assert.Equal(t, "test-cluster", logEntry["cluster"])
	assert.Equal(t, "サービス情報", logEntry["msg"])
}

func TestLoggerFileOutput(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &logger.Config{
		Level:      "info",
		Format:     "json",
		Filename:   logFile,
		MaxSize:    1, // 1MB
		MaxAge:     1, // 1日
		MaxBackups: 3,
	}

	logger, err := logger.NewLogger(config)
	require.NoError(t, err)

	logger.Info("ファイル出力テスト")

	// ファイルが作成されたことを確認
	assert.FileExists(t, logFile)

	// ファイル内容を確認
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	var logEntry map[string]interface{}
	err = json.Unmarshal(content, &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "ファイル出力テスト", logEntry["msg"])
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  "debug",
		Format: "json",
		Output: &buf,
	}

	logger, err := logger.NewLogger(config)
	require.NoError(t, err)

	logger.Debug("デバッグメッセージ")
	logger.Info("インフォメッセージ")
	logger.Warn("警告メッセージ")
	logger.Error("エラーメッセージ")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 4)

	// 各ログレベルを確認
	levels := []string{"debug", "info", "warning", "error"}
	for i, line := range lines {
		var logEntry map[string]interface{}
		err := json.Unmarshal([]byte(line), &logEntry)
		require.NoError(t, err)
		assert.Equal(t, levels[i], logEntry["level"])
	}
}

func TestInvalidLogLevel(t *testing.T) {
	config := &logger.Config{
		Level:  "invalid",
		Format: "json",
	}

	_, err := logger.NewLogger(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "無効なログレベル")
}

func TestDefaultConfig(t *testing.T) {
	config := logger.GetDefaultConfig()

	assert.Equal(t, "info", config.Level)
	assert.Equal(t, "text", config.Format)
	assert.Equal(t, "", config.Filename)
	assert.Equal(t, 100, config.MaxSize)
	assert.Equal(t, 30, config.MaxAge)
	assert.Equal(t, 10, config.MaxBackups)
}
