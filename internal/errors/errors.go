package errors

import (
	"fmt"
)

// ErrorType はエラーの種類を表す
type ErrorType int

const (
	// ErrTypeConfig 設定関連のエラー
	ErrTypeConfig ErrorType = iota
	// ErrTypeAWS AWS API関連のエラー
	ErrTypeAWS
	// ErrTypeValidation バリデーション関連のエラー
	ErrTypeValidation
	// ErrTypeNetwork ネットワーク関連のエラー
	ErrTypeNetwork
	// ErrTypeGeneral 一般的なエラー
	ErrTypeGeneral
)

// PhantomError はphantom-ecs専用のエラー型
type PhantomError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// Error は error インターフェースの実装
func (e *PhantomError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// GetExitCode はエラータイプに基づいて適切な終了コードを返す
func (e *PhantomError) GetExitCode() int {
	switch e.Type {
	case ErrTypeConfig:
		return 1
	case ErrTypeAWS:
		return 2
	case ErrTypeValidation:
		return 3
	case ErrTypeNetwork:
		return 4
	case ErrTypeGeneral:
		return 5
	default:
		return 1
	}
}

// NewPhantomError は新しいPhantomErrorを作成する
func NewPhantomError(errType ErrorType, message string, cause error) *PhantomError {
	return &PhantomError{
		Type:    errType,
		Message: message,
		Cause:   cause,
	}
}

// WrapError は既存のエラーをPhantomErrorでラップする
func WrapError(errType ErrorType, message string, cause error) *PhantomError {
	return NewPhantomError(errType, message, cause)
}

// IsPhantomError は与えられたエラーがPhantomErrorかどうかを判定する
func IsPhantomError(err error) bool {
	_, ok := err.(*PhantomError)
	return ok
}

// エラータイプ別のヘルパー関数

// NewConfigError は設定関連のエラーを作成する
func NewConfigError(message string, cause error) *PhantomError {
	return NewPhantomError(ErrTypeConfig, message, cause)
}

// NewAWSError はAWS関連のエラーを作成する
func NewAWSError(message string, cause error) *PhantomError {
	return NewPhantomError(ErrTypeAWS, message, cause)
}

// NewValidationError はバリデーション関連のエラーを作成する
func NewValidationError(message string, cause error) *PhantomError {
	return NewPhantomError(ErrTypeValidation, message, cause)
}

// NewNetworkError はネットワーク関連のエラーを作成する
func NewNetworkError(message string, cause error) *PhantomError {
	return NewPhantomError(ErrTypeNetwork, message, cause)
}

// NewGeneralError は一般的なエラーを作成する
func NewGeneralError(message string, cause error) *PhantomError {
	return NewPhantomError(ErrTypeGeneral, message, cause)
}

// 定義済みエラーメッセージ
var (
	ErrInvalidRegion          = NewConfigError("無効なリージョンが指定されました", nil)
	ErrConfigFileNotFound     = NewConfigError("設定ファイルが見つかりません", nil)
	ErrInvalidProfile         = NewConfigError("無効なプロファイルが指定されました", nil)
	ErrServiceNotFound        = NewAWSError("指定されたサービスが見つかりません", nil)
	ErrClusterNotFound        = NewAWSError("指定されたクラスターが見つかりません", nil)
	ErrInsufficientPermission = NewAWSError("権限が不足しています", nil)
	ErrNetworkTimeout         = NewNetworkError("ネットワークタイムアウトが発生しました", nil)
	ErrRateLimitExceeded      = NewNetworkError("レート制限に達しました", nil)
)
