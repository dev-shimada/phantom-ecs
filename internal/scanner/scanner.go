package scanner

import (
	"context"
	"strings"

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

// Scanner はECSサービスをスキャンする機能を提供
type Scanner struct {
	client ECSClient
}

// NewScanner は新しいScannerインスタンスを作成
func NewScanner(client ECSClient) *Scanner {
	return &Scanner{
		client: client,
	}
}

// ScanServices は指定されたクラスターからECSサービスを取得
func (s *Scanner) ScanServices(ctx context.Context, clusterNames []string) ([]models.ECSService, error) {
	var allServices []models.ECSService

	for _, clusterName := range clusterNames {
		services, err := s.scanServicesInCluster(ctx, clusterName)
		if err != nil {
			return nil, err
		}
		allServices = append(allServices, services...)
	}

	return allServices, nil
}

// DiscoverClusters は利用可能なクラスターを発見
func (s *Scanner) DiscoverClusters(ctx context.Context) ([]string, error) {
	output, err := s.client.ListClusters(ctx, &ecs.ListClustersInput{})
	if err != nil {
		return nil, err
	}

	var clusterNames []string
	for _, clusterArn := range output.ClusterArns {
		// ARN形式からクラスター名を抽出
		// arn:aws:ecs:region:account:cluster/cluster-name
		parts := strings.Split(clusterArn, "/")
		if len(parts) > 0 {
			clusterNames = append(clusterNames, parts[len(parts)-1])
		}
	}

	return clusterNames, nil
}

// scanServicesInCluster は単一のクラスター内のサービスをスキャン
func (s *Scanner) scanServicesInCluster(ctx context.Context, clusterName string) ([]models.ECSService, error) {
	// サービス一覧を取得
	listOutput, err := s.client.ListServices(ctx, &ecs.ListServicesInput{
		Cluster: &clusterName,
	})
	if err != nil {
		return nil, err
	}

	// サービスがない場合は空のスライスを返す
	if len(listOutput.ServiceArns) == 0 {
		return []models.ECSService{}, nil
	}

	// サービス詳細を取得
	describeOutput, err := s.client.DescribeServices(ctx, &ecs.DescribeServicesInput{
		Cluster:  &clusterName,
		Services: listOutput.ServiceArns,
	})
	if err != nil {
		return nil, err
	}

	// AWS ECSサービス情報をモデルに変換
	var services []models.ECSService
	for _, service := range describeOutput.Services {
		ecsService := s.convertToECSService(service, clusterName)
		services = append(services, ecsService)
	}

	return services, nil
}

// convertToECSService はAWS ECSサービス情報をモデルに変換
func (s *Scanner) convertToECSService(service types.Service, clusterName string) models.ECSService {
	ecsService := models.ECSService{
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

	// CreatedAtとLaunchTypeは現在のテストでは使用されていないため、デフォルト値を設定
	if service.LaunchType != "" {
		ecsService.LaunchType = string(service.LaunchType)
	}

	if service.CreatedAt != nil {
		ecsService.CreatedAt = *service.CreatedAt
	}

	return ecsService
}
