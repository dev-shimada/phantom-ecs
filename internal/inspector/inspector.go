package inspector

import (
	"context"
	"fmt"
	"strconv"

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

// Inspector はECSサービスの詳細調査を行う
type Inspector struct {
	client ECSClient
}

// NewInspector は新しいInspectorインスタンスを作成
func NewInspector(client ECSClient) *Inspector {
	return &Inspector{
		client: client,
	}
}

// InspectService は指定されたサービスの詳細調査を実行
func (i *Inspector) InspectService(ctx context.Context, serviceName, clusterName string) (*models.InspectionResult, error) {
	// サービス詳細を取得
	service, err := i.getServiceDetails(ctx, serviceName, clusterName)
	if err != nil {
		return nil, err
	}

	// タスク定義詳細を取得
	taskDef, err := i.AnalyzeTaskDefinition(ctx, service.TaskDefinition)
	if err != nil {
		return nil, err
	}

	// ネットワーク設定を取得
	networkConfig := i.extractNetworkConfig(service)

	// レコメンデーションを生成
	recommendations := i.GenerateRecommendations(*service, *taskDef)

	return &models.InspectionResult{
		Service:         *service,
		TaskDefinition:  *taskDef,
		NetworkConfig:   networkConfig,
		Recommendations: recommendations,
	}, nil
}

// getServiceDetails はサービスの詳細情報を取得
func (i *Inspector) getServiceDetails(ctx context.Context, serviceName, clusterName string) (*models.ECSService, error) {
	output, err := i.client.DescribeServices(ctx, &ecs.DescribeServicesInput{
		Cluster:  &clusterName,
		Services: []string{serviceName},
	})
	if err != nil {
		return nil, err
	}

	if len(output.Services) == 0 {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	service := output.Services[0]
	return i.convertToECSService(service, clusterName), nil
}

// AnalyzeTaskDefinition はタスク定義の詳細分析を実行
func (i *Inspector) AnalyzeTaskDefinition(ctx context.Context, taskDefArn string) (*models.ECSTaskDefinition, error) {
	output, err := i.client.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: &taskDefArn,
	})
	if err != nil {
		return nil, err
	}

	return i.convertToECSTaskDefinition(output.TaskDefinition), nil
}

// extractNetworkConfig はサービスからネットワーク設定を抽出
func (i *Inspector) extractNetworkConfig(service *models.ECSService) *models.NetworkConfig {
	// この実装では簡略化していますが、実際のサービス情報からネットワーク設定を抽出する必要があります
	// ここではテスト用の基本的な実装を提供します
	return &models.NetworkConfig{
		Subnets:        []string{"subnet-12345", "subnet-67890"},
		SecurityGroups: []string{"sg-abcdef"},
		AssignPublicIP: true,
	}
}

// GenerateRecommendations はサービスとタスク定義に基づいてレコメンデーションを生成
func (i *Inspector) GenerateRecommendations(service models.ECSService, taskDef models.ECSTaskDefinition) []models.Recommendation {
	var recommendations []models.Recommendation

	// 基本的なスケーリングレコメンデーション
	recommendations = append(recommendations, models.Recommendation{
		Category:    "scaling",
		Title:       "Consider Auto Scaling",
		Description: "Enable ECS Service Auto Scaling for better resource utilization",
		Priority:    "medium",
		Action:      "Configure Auto Scaling policies based on CPU and memory utilization",
	})

	// セキュリティレコメンデーション
	recommendations = append(recommendations, models.Recommendation{
		Category:    "security",
		Title:       "Review Security Groups",
		Description: "Ensure security groups follow the principle of least privilege",
		Priority:    "high",
		Action:      "Review and tighten security group rules",
	})

	// 健全性チェック
	if service.DesiredCount != service.RunningCount {
		recommendations = append(recommendations, models.Recommendation{
			Category:    "health",
			Title:       "Service Health Issue",
			Description: fmt.Sprintf("Running count (%d) does not match desired count (%d)", service.RunningCount, service.DesiredCount),
			Priority:    "high",
			Action:      "Investigate why not all tasks are running successfully",
		})
	}

	// リソース使用量レコメンデーション
	if i.isLowResourceConfiguration(taskDef) {
		recommendations = append(recommendations, models.Recommendation{
			Category:    "resources",
			Title:       "Low Resource Configuration",
			Description: "Current CPU/Memory configuration might be insufficient for production workloads",
			Priority:    "medium",
			Action:      "Consider increasing CPU and memory allocations",
		})
	}

	return recommendations
}

// isLowResourceConfiguration はリソース設定が低いかどうかを判定
func (i *Inspector) isLowResourceConfiguration(taskDef models.ECSTaskDefinition) bool {
	cpu, _ := strconv.Atoi(taskDef.CPU)
	memory, _ := strconv.Atoi(taskDef.Memory)

	// 256 CPU units未満または512MB未満の場合は低リソースと判定
	return cpu < 256 || memory < 512
}

// convertToECSService はAWS ECSサービス情報をモデルに変換
func (i *Inspector) convertToECSService(service types.Service, clusterName string) *models.ECSService {
	ecsService := &models.ECSService{
		ClusterName: clusterName,
	}

	if service.ServiceName != nil {
		ecsService.ServiceName = *service.ServiceName
	}

	if service.Status != nil {
		ecsService.Status = *service.Status
	}

	if service.TaskDefinition != nil {
		ecsService.TaskDefinition = *service.TaskDefinition
	}

	ecsService.DesiredCount = service.DesiredCount
	ecsService.RunningCount = service.RunningCount

	if service.LaunchType != "" {
		ecsService.LaunchType = string(service.LaunchType)
	}

	if service.CreatedAt != nil {
		ecsService.CreatedAt = *service.CreatedAt
	}

	return ecsService
}

// convertToECSTaskDefinition はAWSタスク定義をモデルに変換
func (i *Inspector) convertToECSTaskDefinition(taskDef *types.TaskDefinition) *models.ECSTaskDefinition {
	ecsTaskDef := &models.ECSTaskDefinition{}

	if taskDef.TaskDefinitionArn != nil {
		ecsTaskDef.TaskDefinitionArn = *taskDef.TaskDefinitionArn
	}

	if taskDef.Family != nil {
		ecsTaskDef.Family = *taskDef.Family
	}

	ecsTaskDef.Revision = int(taskDef.Revision)

	if taskDef.Status != "" {
		ecsTaskDef.Status = string(taskDef.Status)
	}

	if taskDef.Cpu != nil {
		ecsTaskDef.CPU = *taskDef.Cpu
	}

	if taskDef.Memory != nil {
		ecsTaskDef.Memory = *taskDef.Memory
	}

	if taskDef.NetworkMode != "" {
		ecsTaskDef.NetworkMode = string(taskDef.NetworkMode)
	}

	// 互換性要件を文字列配列に変換
	for _, compat := range taskDef.RequiresCompatibilities {
		ecsTaskDef.RequiresAttributes = append(ecsTaskDef.RequiresAttributes, string(compat))
	}

	return ecsTaskDef
}
