package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dev-shimada/phantom-ecs/internal/models"
	"gopkg.in/yaml.v3"
)

// Formatter は出力フォーマット機能を提供
type Formatter struct{}

// FormatOptions はフォーマットオプションを表す構造体
type FormatOptions struct {
	Format       string `json:"format"`        // json, yaml, table, compact
	PrettyPrint  bool   `json:"pretty_print"`  // プリティプリント有効
	IncludeEmpty bool   `json:"include_empty"` // 空の値を含める
}

// NewFormatter は新しいFormatterインスタンスを作成
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FormatJSON はデータをJSON形式でフォーマット
func (f *Formatter) FormatJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// FormatYAML はデータをYAML形式でフォーマット
func (f *Formatter) FormatYAML(data interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}

// FormatTable はデータをテーブル形式でフォーマット
func (f *Formatter) FormatTable(data interface{}) (string, error) {
	switch v := data.(type) {
	case []models.ECSService:
		return f.formatECSServicesTable(v), nil
	case models.DeploymentResult:
		return f.formatDeploymentResultTable(v), nil
	case models.InspectionResult:
		return f.formatInspectionResultTable(v), nil
	default:
		return "", fmt.Errorf("unsupported data type for table format: %T", data)
	}
}

// FormatCompact はデータをコンパクト形式でフォーマット
func (f *Formatter) FormatCompact(data interface{}) (string, error) {
	switch v := data.(type) {
	case []models.ECSService:
		return f.formatECSServicesCompact(v), nil
	default:
		return "", fmt.Errorf("unsupported data type for compact format: %T", data)
	}
}

// FormatWithOptions は指定されたオプションでデータをフォーマット
func (f *Formatter) FormatWithOptions(data interface{}, options FormatOptions) (string, error) {
	switch options.Format {
	case "json":
		if options.PrettyPrint {
			return f.FormatJSON(data)
		}
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		return string(jsonBytes), nil
	case "yaml":
		return f.FormatYAML(data)
	case "table":
		return f.FormatTable(data)
	case "compact":
		return f.FormatCompact(data)
	default:
		return "", fmt.Errorf("unsupported format: %s", options.Format)
	}
}

// formatECSServicesTable はECSサービス一覧をテーブル形式でフォーマット
func (f *Formatter) formatECSServicesTable(services []models.ECSService) string {
	if len(services) == 0 {
		return "No services found."
	}

	var result strings.Builder

	// ヘッダー
	header := fmt.Sprintf("%-20s %-15s %-10s %-25s %-8s %-8s %-12s",
		"SERVICE NAME", "CLUSTER", "STATUS", "TASK DEFINITION", "DESIRED", "RUNNING", "LAUNCH TYPE")
	result.WriteString(header + "\n")

	// 区切り線
	separator := strings.Repeat("-", len(header))
	result.WriteString(separator + "\n")

	// データ行
	for _, service := range services {
		row := fmt.Sprintf("%-20s %-15s %-10s %-25s %-8d %-8d %-12s",
			f.truncateString(service.ServiceName, 20),
			f.truncateString(service.ClusterName, 15),
			service.Status,
			f.truncateString(service.TaskDefinition, 25),
			service.DesiredCount,
			service.RunningCount,
			service.LaunchType)
		result.WriteString(row + "\n")
	}

	return result.String()
}

// formatDeploymentResultTable はデプロイメント結果をテーブル形式でフォーマット
func (f *Formatter) formatDeploymentResultTable(result models.DeploymentResult) string {
	var output strings.Builder

	header := fmt.Sprintf("%-20s %-15s %-8s %-8s %-50s",
		"SERVICE NAME", "CLUSTER", "SUCCESS", "DRY RUN", "TASK DEFINITION ARN")
	output.WriteString(header + "\n")

	separator := strings.Repeat("-", len(header))
	output.WriteString(separator + "\n")

	row := fmt.Sprintf("%-20s %-15s %-8t %-8t %-50s",
		f.truncateString(result.ServiceName, 20),
		f.truncateString(result.ClusterName, 15),
		result.Success,
		result.DryRun,
		f.truncateString(result.TaskDefinitionArn, 50))
	output.WriteString(row + "\n")

	return output.String()
}

// formatInspectionResultTable はインスペクション結果をテーブル形式でフォーマット
func (f *Formatter) formatInspectionResultTable(result models.InspectionResult) string {
	var output strings.Builder

	output.WriteString("=== SERVICE INFORMATION ===\n")
	output.WriteString(f.formatECSServicesTable([]models.ECSService{result.Service}))

	output.WriteString("\n=== TASK DEFINITION ===\n")
	output.WriteString(fmt.Sprintf("Family: %s\n", result.TaskDefinition.Family))
	output.WriteString(fmt.Sprintf("Revision: %d\n", result.TaskDefinition.Revision))
	output.WriteString(fmt.Sprintf("CPU: %s\n", result.TaskDefinition.CPU))
	output.WriteString(fmt.Sprintf("Memory: %s\n", result.TaskDefinition.Memory))
	output.WriteString(fmt.Sprintf("Network Mode: %s\n", result.TaskDefinition.NetworkMode))

	if result.NetworkConfig != nil {
		output.WriteString("\n=== NETWORK CONFIGURATION ===\n")
		output.WriteString(fmt.Sprintf("Subnets: %s\n", strings.Join(result.NetworkConfig.Subnets, ", ")))
		output.WriteString(fmt.Sprintf("Security Groups: %s\n", strings.Join(result.NetworkConfig.SecurityGroups, ", ")))
		output.WriteString(fmt.Sprintf("Assign Public IP: %t\n", result.NetworkConfig.AssignPublicIP))
	}

	if len(result.Recommendations) > 0 {
		output.WriteString("\n=== RECOMMENDATIONS ===\n")
		for i, rec := range result.Recommendations {
			output.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, strings.ToUpper(rec.Priority), rec.Title))
			output.WriteString(fmt.Sprintf("   Category: %s\n", rec.Category))
			output.WriteString(fmt.Sprintf("   Description: %s\n", rec.Description))
			output.WriteString(fmt.Sprintf("   Action: %s\n", rec.Action))
			output.WriteString("\n")
		}
	}

	return output.String()
}

// formatECSServicesCompact はECSサービス一覧をコンパクト形式でフォーマット
func (f *Formatter) formatECSServicesCompact(services []models.ECSService) string {
	if len(services) == 0 {
		return "No services found."
	}

	var result strings.Builder

	for _, service := range services {
		status := "✓"
		if !f.IsHealthyService(service) {
			status = "✗"
		}

		line := fmt.Sprintf("%s %s/%s [%s] %d/%d %s",
			status,
			service.ClusterName,
			service.ServiceName,
			service.Status,
			service.RunningCount,
			service.DesiredCount,
			service.LaunchType)
		result.WriteString(line + "\n")
	}

	return result.String()
}

// IsHealthyService はサービスが健全状態かどうかを判定
func (f *Formatter) IsHealthyService(service models.ECSService) bool {
	return service.Status == "ACTIVE" && service.DesiredCount == service.RunningCount
}

// truncateString は文字列を指定された長さに切り詰める
func (f *Formatter) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// GetSupportedFormats はサポートされている出力形式一覧を返す
func (f *Formatter) GetSupportedFormats() []string {
	return []string{"json", "yaml", "table", "compact"}
}

// ValidateFormat は指定された形式がサポートされているかチェック
func (f *Formatter) ValidateFormat(format string) bool {
	supportedFormats := f.GetSupportedFormats()
	for _, supported := range supportedFormats {
		if format == supported {
			return true
		}
	}
	return false
}
