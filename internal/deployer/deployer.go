package deployer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/dev-shimada/phantom-ecs/internal/models"
)

// ECSClient はECS操作のインターフェース
type ECSClient interface {
	ListClusters(ctx context.Context, input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error)
	ListServices(ctx context.Context, input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error)
	DescribeServices(ctx context.Context, input *ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error)
	DescribeTaskDefinition(ctx context.Context, input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error)
	CreateService(ctx context.Context, input *ecs.CreateServiceInput) (*ecs.CreateServiceOutput, error)
	RegisterTaskDefinition(ctx context.Context, input *ecs.RegisterTaskDefinitionInput) (*ecs.RegisterTaskDefinitionOutput, error)
}

// DeploymentCustomization はmodelsパッケージから取得
type DeploymentCustomization = models.DeploymentCustomization

// Deployer はECSサービスのデプロイを行う
type Deployer struct {
	client ECSClient
}

// NewDeployer は新しいDeployerインスタンスを作成
func NewDeployer(client ECSClient) *Deployer {
	return &Deployer{
		client: client,
	}
}

// DeployService は指定されたサービスをデプロイする
func (d *Deployer) DeployService(ctx context.Context, inspectionResult *models.InspectionResult, targetCluster, newServiceName string, dryRun bool) (*models.DeploymentResult, error) {
	// バリデーション
	err := d.ValidateDeployment(inspectionResult, targetCluster, newServiceName)
	if err != nil {
		return &models.DeploymentResult{
			ServiceName: newServiceName,
			ClusterName: targetCluster,
			Success:     false,
			DryRun:      dryRun,
			Error:       err.Error(),
		}, err
	}

	var operations []string

	// Dry runの場合は実行せずに予定操作を返す
	if dryRun {
		operations = append(operations, fmt.Sprintf("Register task definition: %s-copy", inspectionResult.TaskDefinition.Family))
		operations = append(operations, fmt.Sprintf("Create service: %s in cluster %s", newServiceName, targetCluster))

		return &models.DeploymentResult{
			ServiceName: newServiceName,
			ClusterName: targetCluster,
			Success:     true,
			DryRun:      true,
			Operations:  operations,
		}, nil
	}

	// タスク定義を複製
	newTaskDefFamily := fmt.Sprintf("%s-copy", inspectionResult.TaskDefinition.Family)
	taskDefArn, err := d.CloneTaskDefinition(ctx, inspectionResult.TaskDefinition, newTaskDefFamily)
	if err != nil {
		return &models.DeploymentResult{
			ServiceName: newServiceName,
			ClusterName: targetCluster,
			Success:     false,
			Error:       fmt.Sprintf("failed to clone task definition: %v", err),
		}, err
	}

	// サービスを作成
	err = d.createService(ctx, inspectionResult, targetCluster, newServiceName, taskDefArn)
	if err != nil {
		return &models.DeploymentResult{
			ServiceName:       newServiceName,
			ClusterName:       targetCluster,
			TaskDefinitionArn: taskDefArn,
			Success:           false,
			Error:             fmt.Sprintf("failed to create service: %v", err),
		}, err
	}

	return &models.DeploymentResult{
		ServiceName:       newServiceName,
		ClusterName:       targetCluster,
		TaskDefinitionArn: taskDefArn,
		Success:           true,
		DryRun:            false,
	}, nil
}

// CloneTaskDefinition はタスク定義を複製する
func (d *Deployer) CloneTaskDefinition(ctx context.Context, sourceTaskDef models.ECSTaskDefinition, newFamily string) (string, error) {
	// タスク定義登録用の入力を作成
	input := &ecs.RegisterTaskDefinitionInput{
		Family:                  &newFamily,
		Cpu:                     &sourceTaskDef.CPU,
		Memory:                  &sourceTaskDef.Memory,
		NetworkMode:             types.NetworkMode(sourceTaskDef.NetworkMode),
		RequiresCompatibilities: []types.Compatibility{},
		ContainerDefinitions: []types.ContainerDefinition{
			// 基本的なコンテナ定義（実際の実装では元のタスク定義から取得）
			{
				Name:  stringPtr("app"),
				Image: stringPtr("nginx:latest"),
			},
		},
	}

	// 互換性要件を変換
	for _, attr := range sourceTaskDef.RequiresAttributes {
		input.RequiresCompatibilities = append(input.RequiresCompatibilities, types.Compatibility(attr))
	}

	// タスク定義を登録
	output, err := d.client.RegisterTaskDefinition(ctx, input)
	if err != nil {
		return "", err
	}

	if output.TaskDefinition.TaskDefinitionArn != nil {
		return *output.TaskDefinition.TaskDefinitionArn, nil
	}

	return "", fmt.Errorf("failed to get task definition ARN")
}

// createService はサービスを作成する
func (d *Deployer) createService(ctx context.Context, inspectionResult *models.InspectionResult, targetCluster, serviceName, taskDefArn string) error {
	input := &ecs.CreateServiceInput{
		ServiceName:    &serviceName,
		Cluster:        &targetCluster,
		TaskDefinition: &taskDefArn,
		DesiredCount:   &inspectionResult.Service.DesiredCount,
		LaunchType:     types.LaunchType(inspectionResult.Service.LaunchType),
	}

	// ネットワーク設定があれば追加
	if inspectionResult.NetworkConfig != nil {
		input.NetworkConfiguration = &types.NetworkConfiguration{
			AwsvpcConfiguration: &types.AwsVpcConfiguration{
				Subnets:        inspectionResult.NetworkConfig.Subnets,
				SecurityGroups: inspectionResult.NetworkConfig.SecurityGroups,
			},
		}

		if inspectionResult.NetworkConfig.AssignPublicIP {
			input.NetworkConfiguration.AwsvpcConfiguration.AssignPublicIp = types.AssignPublicIpEnabled
		} else {
			input.NetworkConfiguration.AwsvpcConfiguration.AssignPublicIp = types.AssignPublicIpDisabled
		}
	}

	_, err := d.client.CreateService(ctx, input)
	return err
}

// CustomizeService はサービス設定をカスタマイズする
func (d *Deployer) CustomizeService(sourceService models.ECSService, customization DeploymentCustomization) models.ECSService {
	result := sourceService

	// 新しいサービス名
	if customization.NewServiceName != "" {
		result.ServiceName = customization.NewServiceName
	}

	// ターゲットクラスター
	if customization.TargetCluster != "" {
		result.ClusterName = customization.TargetCluster
	}

	// 希望するタスク数
	if customization.DesiredCount != nil {
		result.DesiredCount = *customization.DesiredCount
	}

	// 起動タイプ
	if customization.LaunchType != "" {
		result.LaunchType = customization.LaunchType
	}

	return result
}

// ValidateDeployment はデプロイメントの事前バリデーションを行う
func (d *Deployer) ValidateDeployment(inspectionResult *models.InspectionResult, targetCluster, newServiceName string) error {
	// ソースサービスの状態チェック
	if inspectionResult.Service.Status != "ACTIVE" {
		return fmt.Errorf("source service is not active: %s", inspectionResult.Service.Status)
	}

	// タスク定義の状態チェック
	if inspectionResult.TaskDefinition.Status != "ACTIVE" {
		return fmt.Errorf("source task definition is not active: %s", inspectionResult.TaskDefinition.Status)
	}

	// ターゲットクラスター名の検証
	if targetCluster == "" {
		return fmt.Errorf("target cluster name cannot be empty")
	}

	// 新しいサービス名の検証
	if newServiceName == "" {
		return fmt.Errorf("new service name cannot be empty")
	}

	// ソースと同じ名前は避ける
	if inspectionResult.Service.ServiceName == newServiceName && inspectionResult.Service.ClusterName == targetCluster {
		return fmt.Errorf("cannot deploy to the same service name in the same cluster")
	}

	return nil
}

// ヘルパー関数
func stringPtr(s string) *string {
	return &s
}
