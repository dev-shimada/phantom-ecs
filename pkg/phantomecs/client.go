package phantomecs

import (
	"context"

	"github.com/dev-shimada/phantom-ecs/internal/aws"
	"github.com/dev-shimada/phantom-ecs/internal/config"
)

// PhantomECSClient phantom-ecsの公開クライアント
type PhantomECSClient struct {
	awsClient  *aws.Client
	ecsService *aws.ECSService
	config     *config.Config
}

// NewPhantomECSClient 新しいPhantomECSクライアントを作成
func NewPhantomECSClient(ctx context.Context, region, profile string) (*PhantomECSClient, error) {
	// AWS クライアントの作成
	awsClient, err := aws.NewClient(ctx, region, profile)
	if err != nil {
		return nil, err
	}

	// 設定の作成
	cfg := config.NewConfig(region, profile)

	// ECSサービスの作成
	ecsService := aws.NewECSService(awsClient)

	return &PhantomECSClient{
		awsClient:  awsClient,
		ecsService: ecsService,
		config:     cfg,
	}, nil
}

// GetConfig 設定を取得
func (p *PhantomECSClient) GetConfig() *config.Config {
	return p.config
}

// GetECSService ECSサービスを取得
func (p *PhantomECSClient) GetECSService() *aws.ECSService {
	return p.ecsService
}

// GetAWSClient AWSクライアントを取得
func (p *PhantomECSClient) GetAWSClient() *aws.Client {
	return p.awsClient
}
