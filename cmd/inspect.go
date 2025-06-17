package cmd

import (
	"context"
	"fmt"

	"github.com/dev-shimada/phantom-ecs/internal/aws"
	"github.com/dev-shimada/phantom-ecs/internal/inspector"
	"github.com/dev-shimada/phantom-ecs/internal/models"
	"github.com/dev-shimada/phantom-ecs/internal/utils"
	"github.com/spf13/cobra"
)

// InspectorInterface はInspectorの操作を定義するインターフェース
type InspectorInterface interface {
	InspectService(ctx context.Context, serviceName, clusterName string) (*models.InspectionResult, error)
}

// NewInspectCommand はinspectコマンドを作成
func NewInspectCommand(inspectorImpl InspectorInterface) *cobra.Command {
	var clusterName string
	var outputFormat string
	var region string
	var profile string

	cmd := &cobra.Command{
		Use:   "inspect <service-name>",
		Short: "指定されたECSサービスの詳細情報を表示",
		Long: `指定されたECSサービスの詳細情報を表示します。

サービスの基本情報、タスク定義、ネットワーク設定、
レコメンデーションを含む包括的な分析結果を提供します。`,
		Example: `  # 基本的なサービス検査
  phantom-ecs inspect my-service --cluster my-cluster

  # JSON形式で出力
  phantom-ecs inspect my-service --cluster my-cluster --output json

  # 特定のリージョンとプロファイルを使用
  phantom-ecs inspect my-service --cluster my-cluster --region us-west-2 --profile production`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]
			return runInspect(cmd, inspectorImpl, serviceName, clusterName, outputFormat, region, profile)
		},
	}

	// ローカルフラグを定義
	cmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "クラスター名 (必須)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "出力形式 (json|yaml|table)")
	cmd.Flags().StringVarP(&region, "region", "r", "us-east-1", "AWSリージョン")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "AWSプロファイル")

	// 必須フラグを設定
	cmd.MarkFlagRequired("cluster")

	return cmd
}

// NewInspectCommandWithDefaults はデフォルトのInspectorでinspectコマンドを作成
func NewInspectCommandWithDefaults() *cobra.Command {
	return NewInspectCommand(nil) // 実際の実装では適切なInspectorを渡す
}

// runInspect はinspectコマンドの実行ロジック
func runInspect(cmd *cobra.Command, inspectorImpl InspectorInterface, serviceName, clusterName, outputFormat, region, profile string) error {
	ctx := context.Background()

	// 必須パラメータの検証
	if serviceName == "" {
		return fmt.Errorf("service name is required")
	}
	if clusterName == "" {
		return fmt.Errorf("cluster name is required")
	}

	// 出力形式の検証
	formatter := utils.NewFormatter()
	if !formatter.ValidateFormat(outputFormat) {
		return fmt.Errorf("unsupported output format: %s. Supported formats: %v",
			outputFormat, formatter.GetSupportedFormats())
	}

	// Inspectorがnilの場合（実際のAWS呼び出し用）は、AWS Inspectorを作成
	var inspectorToUse InspectorInterface
	if inspectorImpl != nil {
		inspectorToUse = inspectorImpl
	} else {
		// 実際のAWS呼び出し用の実装
		awsClient, err := aws.NewClient(ctx, region, profile)
		if err != nil {
			return fmt.Errorf("failed to create AWS client: %w", err)
		}
		inspectorToUse = inspector.NewInspector(awsClient)
	}

	// サービスの詳細調査を実行
	result, err := inspectorToUse.InspectService(ctx, serviceName, clusterName)
	if err != nil {
		return fmt.Errorf("failed to inspect service: %w", err)
	}

	// 結果をフォーマットして出力
	output, err := formatter.FormatWithOptions(*result, utils.FormatOptions{
		Format:      outputFormat,
		PrettyPrint: true,
	})
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Print(output)
	return nil
}
