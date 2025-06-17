package cmd

import (
	"fmt"
	"os"

	"github.com/dev-shimada/phantom-ecs/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	region       string
	profile      string
	outputFormat string
)

// Version はアプリケーションのバージョン
const Version = "1.0.0"

// NewRootCommand はルートコマンドを作成
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "phantom-ecs",
		Short: "AWS ECS サービス調査CLIツール",
		Long: `phantom-ecs は AWS ECS サービスの調査、分析、デプロイを行うCLIツールです。

主な機能:
	 - ECSサービス一覧表示 (scan)
	 - 特定サービスの詳細調査 (inspect)
	 - 同等サービスの自動作成 (deploy)

例:
	 phantom-ecs scan --region us-east-1 --output json
	 phantom-ecs inspect my-service --cluster my-cluster
	 phantom-ecs deploy my-service --target-cluster new-cluster --dry-run`,
		Version: Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			// フラグから値を取得してバリデーション
			regionFlag, _ := cmd.PersistentFlags().GetString("region")
			profileFlag, _ := cmd.PersistentFlags().GetString("profile")
			outputFlag, _ := cmd.PersistentFlags().GetString("output")

			// 設定を作成してバリデーション
			cfg := config.NewConfig(regionFlag, profileFlag)
			cfg.SetOutputFormat(outputFlag)

			if err := cfg.Validate(); err != nil {
				return err
			}

			// サブコマンドが指定されていない場合はヘルプを表示
			return cmd.Help()
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
	}

	// グローバルフラグを定義
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "設定ファイルパス (default: $HOME/.phantom-ecs.yaml)")
	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "us-east-1", "AWSリージョン")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "AWSプロファイル")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "出力形式 (json|yaml|table)")

	// Viperでフラグをバインド
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))

	// サブコマンドを追加
	rootCmd.AddCommand(NewScanCommandWithDefaults())
	rootCmd.AddCommand(NewInspectCommandWithDefaults())
	rootCmd.AddCommand(NewDeployCommandWithDefaults())
	rootCmd.AddCommand(NewBatchCommand())

	return rootCmd
}

// Execute はルートコマンドを実行
func Execute() {
	rootCmd := NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// initConfig は設定を初期化
func initConfig() error {
	if cfgFile != "" {
		// 設定ファイルが指定された場合
		viper.SetConfigFile(cfgFile)
	} else {
		// デフォルトの設定ファイルを検索
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".phantom-ecs")
		viper.SetConfigType("yaml")
	}

	// 環境変数からの読み込み
	viper.AutomaticEnv()
	viper.SetEnvPrefix("PHANTOM_ECS")

	// 設定ファイルを読み込み（存在しない場合はエラーにしない）
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// 設定の検証
	cfg := config.NewConfig(viper.GetString("region"), viper.GetString("profile"))
	cfg.SetOutputFormat(viper.GetString("output"))

	return cfg.Validate()
}

// GetConfig は現在の設定を取得
func GetConfig() *config.Config {
	cfg := config.NewConfig(viper.GetString("region"), viper.GetString("profile"))
	cfg.SetOutputFormat(viper.GetString("output"))
	return cfg
}
