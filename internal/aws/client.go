package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// Client AWS操作用のクライアント
type Client struct {
	ecsClient *ecs.Client
	region    string
}

// NewClient 新しいAWSクライアントを作成
func NewClient(ctx context.Context, region, profile string) (*Client, error) {
	// デフォルトリージョンの設定
	if region == "" {
		region = "us-east-1"
	}

	// AWS設定の読み込み
	var cfg aws.Config
	var err error

	if profile != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithSharedConfigProfile(profile),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
		)
	}

	if err != nil {
		return nil, err
	}

	// ECSクライアントの作成
	ecsClient := ecs.NewFromConfig(cfg)

	return &Client{
		ecsClient: ecsClient,
		region:    region,
	}, nil
}

// GetECSClient ECSクライアントを取得
func (c *Client) GetECSClient() *ecs.Client {
	return c.ecsClient
}

// GetRegion 設定されたリージョンを取得
func (c *Client) GetRegion() string {
	return c.region
}

// scanner.ECSClientインターフェースの実装
func (c *Client) ListClusters(ctx context.Context, input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
	return c.ecsClient.ListClusters(ctx, input)
}

func (c *Client) ListServices(ctx context.Context, input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
	return c.ecsClient.ListServices(ctx, input)
}

func (c *Client) DescribeServices(ctx context.Context, input *ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error) {
	return c.ecsClient.DescribeServices(ctx, input)
}

func (c *Client) DescribeTaskDefinition(ctx context.Context, input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
	return c.ecsClient.DescribeTaskDefinition(ctx, input)
}

func (c *Client) CreateService(ctx context.Context, input *ecs.CreateServiceInput) (*ecs.CreateServiceOutput, error) {
	return c.ecsClient.CreateService(ctx, input)
}

func (c *Client) RegisterTaskDefinition(ctx context.Context, input *ecs.RegisterTaskDefinitionInput) (*ecs.RegisterTaskDefinitionOutput, error) {
	return c.ecsClient.RegisterTaskDefinition(ctx, input)
}
