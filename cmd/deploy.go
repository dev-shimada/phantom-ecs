package cmd

import (
	"context"
	"fmt"

	"github.com/dev-shimada/phantom-ecs/internal/aws"
	"github.com/dev-shimada/phantom-ecs/internal/deployer"
	"github.com/dev-shimada/phantom-ecs/internal/inspector"
	"github.com/dev-shimada/phantom-ecs/internal/models"
	"github.com/dev-shimada/phantom-ecs/internal/utils"
	"github.com/spf13/cobra"
)

// DeployerInterface はDeployerの操作を定義するインターフェース
type DeployerInterface interface {
	DeployService(ctx context.Context, inspectionResult *models.InspectionResult, targetCluster, newServiceName string, dryRun bool) (*models.DeploymentResult, error)
}

// NewDeployCommand はdeployコマンドを作成
func NewDeployCommand(deployerImpl DeployerInterface, inspectorImpl InspectorInterface) *cobra.Command {
	var fromCluster string
	var targetCluster string
	var newServiceName string
	var dryRun bool
	var outputFormat string
	var region string
	var profile string

	cmd := &cobra.Command{
		Use:   "deploy <service-name>",
		Short: "指定されたECSサービスと同等のサービスを作成",
		Long: `指定されたECSサービスと同等のサービスを作成します。

元のサービスを詳細調査し、その設定を基に新しいクラスターに
同じ構成のサービスを作成します。dry-runモードで事前に
実行内容を確認することができます。`,
		Example: `  # ドライランでデプロイ内容を確認
  phantom-ecs deploy my-service --from-cluster source-cluster --target-cluster target-cluster --dry-run

  # 実際にサービスをデプロイ
  phantom-ecs deploy my-service --from-cluster prod-cluster --target-cluster staging-cluster

  # 新しいサービス名を指定してデプロイ
  phantom-ecs deploy my-service --from-cluster prod-cluster --target-cluster dev-cluster --new-service-name dev-my-service

  # 特定のリージョンとプロファイルを使用
  phantom-ecs deploy my-service --from-cluster source --target-cluster target --region us-west-2 --profile production`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]
			return runDeploy(cmd, deployerImpl, inspectorImpl, serviceName, fromCluster, targetCluster, newServiceName, dryRun, outputFormat, region, profile)
		},
	}

	// ローカルフラグを定義
	cmd.Flags().StringVar(&fromCluster, "from-cluster", "", "コピー元のクラスター名 (必須)")
	cmd.Flags().StringVar(&targetCluster, "target-cluster", "", "デプロイ先のクラスター名 (必須)")
	cmd.Flags().StringVar(&newServiceName, "new-service-name", "", "新しいサービス名 (未指定時は元のサービス名を使用)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "実際には実行せずに処理内容を表示")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "出力形式 (json|yaml|table)")
	cmd.Flags().StringVarP(&region, "region", "r", "us-east-1", "AWSリージョン")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "AWSプロファイル")

	// 必須フラグを設定
	cmd.MarkFlagRequired("from-cluster")
	cmd.MarkFlagRequired("target-cluster")

	return cmd
}

// NewDeployCommandWithDefaults はデフォルトのDeployerとInspectorでdeployコマンドを作成
func NewDeployCommandWithDefaults() *cobra.Command {
	return NewDeployCommand(nil, nil) // 実際の実装では適切なDeployerとInspectorを渡す
}

// runDeploy はdeployコマンドの実行ロジック
func runDeploy(cmd *cobra.Command, deployerImpl DeployerInterface, inspectorImpl InspectorInterface, serviceName, fromCluster, targetCluster, newServiceName string, dryRun bool, outputFormat, region, profile string) error {
	ctx := context.Background()

	// 必須パラメータの検証
	if serviceName == "" {
		return fmt.Errorf("service name is required")
	}
	if fromCluster == "" {
		return fmt.Errorf("from-cluster is required")
	}
	if targetCluster == "" {
		return fmt.Errorf("target-cluster is required")
	}

	// 新しいサービス名のデフォルト設定
	if newServiceName == "" {
		newServiceName = serviceName
	}

	// 出力形式の検証
	formatter := utils.NewFormatter()
	if !formatter.ValidateFormat(outputFormat) {
		return fmt.Errorf("unsupported output format: %s. Supported formats: %v",
			outputFormat, formatter.GetSupportedFormats())
	}

	// DeployerとInspectorがnilの場合（実際のAWS呼び出し用）は、AWS実装を作成
	var deployerToUse DeployerInterface
	var inspectorToUse InspectorInterface

	if deployerImpl != nil && inspectorImpl != nil {
		deployerToUse = deployerImpl
		inspectorToUse = inspectorImpl
	} else {
		// 実際のAWS呼び出し用の実装
		awsClient, err := aws.NewClient(ctx, region, profile)
		if err != nil {
			return fmt.Errorf("failed to create AWS client: %w", err)
		}
		deployerToUse = deployer.NewDeployer(awsClient)
		inspectorToUse = inspector.NewInspector(awsClient)
	}

	// ソースサービスの詳細調査を実行
	inspectionResult, err := inspectorToUse.InspectService(ctx, serviceName, fromCluster)
	if err != nil {
		return fmt.Errorf("failed to inspect source service: %w", err)
	}

	// サービスのデプロイを実行
	deploymentResult, err := deployerToUse.DeployService(ctx, inspectionResult, targetCluster, newServiceName, dryRun)
	if err != nil {
		return fmt.Errorf("failed to deploy service: %w", err)
	}

	// 結果をフォーマットして出力
	output, err := formatter.FormatWithOptions(*deploymentResult, utils.FormatOptions{
		Format:      outputFormat,
		PrettyPrint: true,
	})
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Print(output)
	return nil
}
