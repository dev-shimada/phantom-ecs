# phantom-ecs

AWS ECS サービス調査CLIツール

[![Go Version](https://img.shields.io/badge/Go-1.24.3-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 📖 概要

phantom-ecs は AWS ECS サービスの調査、分析、デプロイを効率的に行うためのCLIツールです。テスト駆動開発（TDD）によって開発され、本番環境での使用を想定した堅牢性を持っています。

### 主な機能

- **🔍 スキャン**: AWS上のECSサービス一覧表示
- **🔎 調査**: 特定ECSサービスの詳細情報取得
- **🚀 デプロイ**: 既存サービスを基にした新しいサービスの作成
- **⚡ バッチ処理**: 複数サービスの同時処理
- **📊 ログ**: 構造化ログとファイルローテーション
- **⚙️ 設定管理**: YAML設定ファイルと環境変数サポート
- **🔄 リトライ**: 自動リトライとレート制限対応

## 🚀 インストール

### バイナリのダウンロード

最新のリリースから実行可能ファイルをダウンロード:

```bash
# Linux/macOS
curl -L https://github.com/dev-shimada/phantom-ecs/releases/latest/download/phantom-ecs-$(uname -s)-$(uname -m) -o phantom-ecs
chmod +x phantom-ecs
sudo mv phantom-ecs /usr/local/bin/
```

### ソースからビルド

```bash
git clone https://github.com/dev-shimada/phantom-ecs.git
cd phantom-ecs
make build
```

### Go installでインストール

```bash
go install github.com/dev-shimada/phantom-ecs@latest
```

## 📋 使用方法

### 基本的なコマンド

#### サービス一覧の表示

```bash
# 基本的なスキャン
phantom-ecs scan

# 特定リージョンでのスキャン
phantom-ecs scan --region ap-northeast-1

# JSON形式での出力
phantom-ecs scan --output json

# 特定プロファイルの使用
phantom-ecs scan --profile production
```

#### サービスの詳細調査

```bash
# サービスの詳細情報を表示
phantom-ecs inspect my-service

# YAML形式での出力
phantom-ecs inspect my-service --output yaml

# 特定クラスターのサービス調査
phantom-ecs inspect my-service --cluster my-cluster
```

#### サービスのデプロイ

```bash
# 既存サービスのコピーを作成
phantom-ecs deploy my-service --target-cluster new-cluster

# Dry runモード（実行せず確認のみ）
phantom-ecs deploy my-service --target-cluster new-cluster --dry-run
```

#### バッチ処理

```bash
# 複数サービスの同時処理
phantom-ecs batch --services service1,service2,service3

# 設定ファイルを使用したバッチ処理
phantom-ecs batch --config-file batch-config.yaml

# 同時実行数とリトライ設定
phantom-ecs batch --services service1,service2 --concurrency 5 --retry-count 3
```

### 設定ファイル

#### YAML設定ファイルの例

```yaml
# ~/.phantom-ecs.yaml
profiles:
  default:
    region: us-east-1
    output_format: table
    
  production:
    region: ap-northeast-1
    output_format: json
    aws_profile: prod-profile
    
  development:
    region: us-west-2
    output_format: yaml

logging:
  level: info
  format: json
  filename: /var/log/phantom-ecs.log
  max_size: 100    # MB
  max_age: 30      # 日
  max_backups: 10  # ファイル数

batch:
  max_concurrency: 5
  retry_attempts: 3
  retry_delay: 2s
  show_progress: true
```

#### 環境変数

```bash
# AWS設定
export AWS_REGION=ap-northeast-1
export AWS_PROFILE=production

# phantom-ecs設定
export PHANTOM_ECS_REGION=ap-northeast-1
export PHANTOM_ECS_OUTPUT_FORMAT=json
export PHANTOM_ECS_LOG_LEVEL=debug
export PHANTOM_ECS_BATCH_MAX_CONCURRENCY=10
```

### コマンドオプション

#### グローバルオプション

- `--region, -r`: AWSリージョン（デフォルト: us-east-1）
- `--profile, -p`: AWSプロファイル
- `--output, -o`: 出力形式（json|yaml|table）
- `--config`: 設定ファイルパス

#### scanコマンド

```bash
phantom-ecs scan [flags]

Flags:
  --region string     AWSリージョン (default "us-east-1")
  --profile string    AWSプロファイル
  --output string     出力形式 (json|yaml|table) (default "table")
```

#### inspectコマンド

```bash
phantom-ecs inspect <service-name> [flags]

Flags:
  --cluster string    クラスター名
  --region string     AWSリージョン (default "us-east-1")
  --profile string    AWSプロファイル
  --output string     出力形式 (json|yaml|table) (default "table")
```

#### deployコマンド

```bash
phantom-ecs deploy <service-name> [flags]

Flags:
  --target-cluster string  作成先クラスター名
  --region string          AWSリージョン (default "us-east-1")
  --profile string         AWSプロファイル
  --dry-run               実行せずに処理内容を表示
```

#### batchコマンド

```bash
phantom-ecs batch [flags]

Flags:
  --services strings       処理対象のサービス名（カンマ区切り）
  --config-file string     バッチ設定ファイルのパス
  --batch-profile string   使用するバッチプロファイル (default "default")
  --concurrency int        同時実行数 (default 3)
  --retry-count int        リトライ回数 (default 3)
  --retry-delay duration   リトライ間隔 (default 2s)
  --progress               プログレスバーを表示 (default true)
  --dry-run               実際には実行せず、処理内容のみ表示
```

## 🔧 開発

### 前提条件

- Go 1.24.3以上
- AWS CLI設定済み
- Docker（テスト用のLocalStack実行時）

### ビルド

```bash
# 開発用ビルド
make build

# リリース用ビルド
make build-release

# 全プラットフォーム向けビルド
make build-all
```

### テスト

```bash
# 単体テスト
make test

# 統合テスト
make test-integration

# カバレッジ付きテスト
make test-coverage

# 全てのテスト
make test-all
```

### ローカル開発

```bash
# 依存関係のインストール
go mod download

# 開発用実行
go run main.go scan --region us-east-1

# テスト実行
go test ./...

# ベンチマーク
go test -bench=. ./...
```

## 🧪 テスト

このプロジェクトはテスト駆動開発（TDD）で開発されています。

### テスト種別

- **単体テスト**: 各パッケージの個別機能テスト
- **統合テスト**: AWS APIとの実際の連携テスト
- **エンドツーエンドテスト**: CLIコマンド実行テスト

### テスト実行

```bash
# 全テスト実行
make test-all

# 特定パッケージのテスト
go test ./internal/scanner/

# 詳細出力でテスト
go test -v ./...

# カバレッジレポート生成
make coverage-html
```

## 📊 パフォーマンス

### ベンチマーク結果

- **バッチ処理**: 100サービスを10並列で約1秒で処理
- **同時実行**: 最大20の同時接続をサポート
- **メモリ使用量**: 通常動作時50MB以下

### 最適化のポイント

- ゴルーチンによる並列処理
- 適切なレート制限
- 効率的なメモリ管理
- コネクションプーリング

## 🛠️ アーキテクチャ

### プロジェクト構造

```
phantom-ecs/
├── cmd/                    # CLIコマンド定義
├── internal/               # 内部パッケージ
│   ├── aws/               # AWS操作
│   ├── batch/             # バッチ処理
│   ├── config/            # 設定管理
│   ├── errors/            # エラーハンドリング
│   ├── logger/            # ロギング
│   ├── models/            # データモデル
│   ├── scanner/           # サービススキャン
│   ├── inspector/         # サービス調査
│   ├── deployer/          # サービスデプロイ
│   └── utils/             # ユーティリティ
├── pkg/                   # 公開パッケージ
├── tests/                 # テスト
├── testdata/              # テストデータ
└── docs/                  # ドキュメント
```

### 設計原則

- **Single Responsibility**: 各パッケージは単一の責務を持つ
- **Dependency Injection**: インターフェースベースの設計
- **Error Handling**: 適切なエラー分類と処理
- **Testability**: テスト可能な設計

## 📝 ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 🤝 コントリビューション

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

### 開発ガイドライン

- テストファーストでの開発
- ゴランの標準コーディング規約に従う
- 適切なドキュメンテーション
- パフォーマンスを考慮した実装

## 📞 サポート

- **Issues**: [GitHub Issues](https://github.com/dev-shimada/phantom-ecs/issues)
- **Discussions**: [GitHub Discussions](https://github.com/dev-shimada/phantom-ecs/discussions)
- **Documentation**: [Wiki](https://github.com/dev-shimada/phantom-ecs/wiki)

## 🗓️ ロードマップ

- [ ] v1.1.0: Fargate Spot対応
- [ ] v1.2.0: ECS Execサポート
- [ ] v1.3.0: CloudFormation統合
- [ ] v2.0.0: EKS対応

---

**作成者**: dev-shimada  
**最終更新**: 2025年6月17日
