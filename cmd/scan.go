package cmd

import (
	"context"
	"fmt"

	"github.com/dev-shimada/phantom-ecs/internal/aws"
	"github.com/dev-shimada/phantom-ecs/internal/models"
	"github.com/dev-shimada/phantom-ecs/internal/scanner"
	"github.com/dev-shimada/phantom-ecs/internal/utils"
	"github.com/spf13/cobra"
)

// ScannerInterface はScannerの操作を定義するインターフェース
type ScannerInterface interface {
	ScanServices(ctx context.Context, clusterNames []string) ([]models.ECSService, error)
	DiscoverClusters(ctx context.Context) ([]string, error)
}

// NewScanCommand はscanコマンドを作成
func NewScanCommand(scannerImpl ScannerInterface) *cobra.Command {
	var outputFormat string
	var region string
	var profile string

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "AWS ECSサービス一覧を表示",
		Long: `AWS ECSサービス一覧を表示します。

指定されたリージョンとプロファイルを使用して、
利用可能なすべてのECSクラスター内のサービスをスキャンし、
指定された形式で結果を出力します。`,
		Example: `  # デフォルト設定でサービス一覧を表示
  phantom-ecs scan

  # 特定のリージョンでスキャン
  phantom-ecs scan --region us-west-2

  # JSON形式で出力
  phantom-ecs scan --output json

  # 特定のプロファイルを使用
  phantom-ecs scan --profile production`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScan(cmd, scannerImpl, outputFormat, region, profile)
		},
	}

	// ローカルフラグを定義
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "出力形式 (json|yaml|table)")
	cmd.Flags().StringVarP(&region, "region", "r", "us-east-1", "AWSリージョン")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "AWSプロファイル")

	return cmd
}

// NewScanCommandWithDefaults はデフォルトのScannerでscanコマンドを作成
func NewScanCommandWithDefaults() *cobra.Command {
	return NewScanCommand(nil) // 実際の実装では適切なScannerを渡す
}

// runScan はscanコマンドの実行ロジック
func runScan(cmd *cobra.Command, scannerImpl ScannerInterface, outputFormat, region, profile string) error {
	ctx := context.Background()

	// 出力形式の検証
	formatter := utils.NewFormatter()
	if !formatter.ValidateFormat(outputFormat) {
		return fmt.Errorf("unsupported output format: %s. Supported formats: %v",
			outputFormat, formatter.GetSupportedFormats())
	}

	// Scannerがnilの場合（実際のAWS呼び出し用）は、AWS Scannerを作成
	var scannerToUse ScannerInterface
	if scannerImpl != nil {
		scannerToUse = scannerImpl
	} else {
		// 実際のAWS呼び出し用の実装
		awsClient, err := aws.NewClient(ctx, region, profile)
		if err != nil {
			return fmt.Errorf("failed to create AWS client: %w", err)
		}
		scannerToUse = scanner.NewScanner(awsClient)
	}

	// クラスターを発見
	clusters, err := scannerToUse.DiscoverClusters(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover clusters: %w", err)
	}

	if len(clusters) == 0 {
		fmt.Println("No ECS clusters found in the specified region.")
		return nil
	}

	// サービスをスキャン
	services, err := scannerToUse.ScanServices(ctx, clusters)
	if err != nil {
		return fmt.Errorf("failed to scan services: %w", err)
	}

	// 結果をフォーマットして出力
	output, err := formatter.FormatWithOptions(services, utils.FormatOptions{
		Format:      outputFormat,
		PrettyPrint: true,
	})
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Print(output)
	return nil
}
