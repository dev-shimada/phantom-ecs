package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config はロガーの設定
type Config struct {
	// Level はログレベル (debug, info, warn, error)
	Level string
	// Format はログフォーマット (json, text)
	Format string
	// Filename はログファイルのパス（空の場合は標準出力）
	Filename string
	// MaxSize はログファイルの最大サイズ（MB）
	MaxSize int
	// MaxAge は保持期間（日）
	MaxAge int
	// MaxBackups は保持するバックアップファイル数
	MaxBackups int
	// Output はカスタム出力先（テスト用）
	Output io.Writer
}

// Logger はロガーのインターフェース
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	WithFields(fields logrus.Fields) *logrus.Entry
	GetLevel() logrus.Level
}

// PhantomLogger はphantom-ecs用のロガー実装
type PhantomLogger struct {
	*logrus.Logger
}

// NewLogger は新しいロガーを作成する
func NewLogger(config *Config) (Logger, error) {
	logger := logrus.New()

	// ログレベル設定
	level, err := parseLogLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("無効なログレベル '%s': %w", config.Level, err)
	}
	logger.SetLevel(level)

	// フォーマット設定
	switch config.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 出力先設定
	if config.Output != nil {
		// テスト用のカスタム出力
		logger.SetOutput(config.Output)
	} else if config.Filename != "" {
		// ファイル出力（ローテーション付き）
		logger.SetOutput(&lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSize,    // MB
			MaxAge:     config.MaxAge,     // 日
			MaxBackups: config.MaxBackups, // ファイル数
			Compress:   true,              // 圧縮
		})
	} else {
		// 標準出力
		logger.SetOutput(os.Stdout)
	}

	return &PhantomLogger{Logger: logger}, nil
}

// parseLogLevel は文字列からログレベルを解析する
func parseLogLevel(level string) (logrus.Level, error) {
	switch level {
	case "debug":
		return logrus.DebugLevel, nil
	case "info":
		return logrus.InfoLevel, nil
	case "warn", "warning":
		return logrus.WarnLevel, nil
	case "error":
		return logrus.ErrorLevel, nil
	default:
		return logrus.InfoLevel, fmt.Errorf("不明なログレベル: %s", level)
	}
}

// GetDefaultConfig はデフォルト設定を返す
func GetDefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Format:     "text",
		Filename:   "",
		MaxSize:    100, // 100MB
		MaxAge:     30,  // 30日
		MaxBackups: 10,  // 10ファイル
	}
}

// NewDefaultLogger はデフォルト設定でロガーを作成する
func NewDefaultLogger() (Logger, error) {
	return NewLogger(GetDefaultConfig())
}

// WithServiceContext はサービス情報を含むロガーを作成する
func (l *PhantomLogger) WithServiceContext(serviceName, clusterName, region string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"service": serviceName,
		"cluster": clusterName,
		"region":  region,
	})
}

// WithOperationContext は操作情報を含むロガーを作成する
func (l *PhantomLogger) WithOperationContext(operation string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"operation": operation,
	})
}

// WithErrorContext はエラー情報を含むロガーを作成する
func (l *PhantomLogger) WithErrorContext(err error) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"error": err.Error(),
	})
}

// LogAWSAPICall はAWS API呼び出しをログに記録する
func (l *PhantomLogger) LogAWSAPICall(service, operation string, duration int64) {
	l.WithFields(logrus.Fields{
		"aws_service":   service,
		"aws_operation": operation,
		"duration_ms":   duration,
	}).Info("AWS API呼び出し完了")
}

// LogProgress は進行状況をログに記録する
func (l *PhantomLogger) LogProgress(current, total int, message string) {
	l.WithFields(logrus.Fields{
		"current": current,
		"total":   total,
		"percent": float64(current) / float64(total) * 100,
	}).Info(message)
}

// 便利な定数
const (
	OperationScan    = "scan"
	OperationInspect = "inspect"
	OperationDeploy  = "deploy"
	OperationBatch   = "batch"
)
