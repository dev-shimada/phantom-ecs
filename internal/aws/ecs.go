package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// ECSServiceInterface ECS操作のインターフェース
type ECSServiceInterface interface {
	ListServices(ctx context.Context, clusterName string) ([]string, error)
	DescribeServices(ctx context.Context, clusterName string, serviceNames []string) (map[string]interface{}, error)
	DescribeTaskDefinition(ctx context.Context, taskDefArn string) (map[string]interface{}, error)
}

// ECSService ECS操作を行うサービス
type ECSService struct {
	client *Client
}

// NewECSService 新しいECSServiceインスタンスを作成
func NewECSService(client *Client) *ECSService {
	return &ECSService{
		client: client,
	}
}

// GetClient クライアントを取得
func (e *ECSService) GetClient() *Client {
	return e.client
}

// ListServices 指定されたクラスターのサービス一覧を取得
func (e *ECSService) ListServices(ctx context.Context, clusterName string) ([]string, error) {
	ecsClient := e.client.GetECSClient()

	input := &ecs.ListServicesInput{}
	if clusterName != "" {
		input.Cluster = &clusterName
	}

	output, err := ecsClient.ListServices(ctx, input)
	if err != nil {
		return nil, err
	}

	var services []string
	for _, arn := range output.ServiceArns {
		services = append(services, arn)
	}

	return services, nil
}

// DescribeServices 指定されたサービスの詳細情報を取得
func (e *ECSService) DescribeServices(ctx context.Context, clusterName string, serviceNames []string) (map[string]interface{}, error) {
	ecsClient := e.client.GetECSClient()

	input := &ecs.DescribeServicesInput{
		Services: serviceNames,
	}
	if clusterName != "" {
		input.Cluster = &clusterName
	}

	output, err := ecsClient.DescribeServices(ctx, input)
	if err != nil {
		return nil, err
	}

	// レスポンスをマップに変換（テスト用）
	result := map[string]interface{}{
		"services": output.Services,
		"failures": output.Failures,
	}

	return result, nil
}

// DescribeTaskDefinition 指定されたタスク定義の詳細情報を取得
func (e *ECSService) DescribeTaskDefinition(ctx context.Context, taskDefArn string) (map[string]interface{}, error) {
	ecsClient := e.client.GetECSClient()

	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: &taskDefArn,
	}

	output, err := ecsClient.DescribeTaskDefinition(ctx, input)
	if err != nil {
		return nil, err
	}

	// レスポンスをマップに変換（テスト用）
	result := map[string]interface{}{
		"taskDefinition": output.TaskDefinition,
		"tags":           output.Tags,
	}

	return result, nil
}
