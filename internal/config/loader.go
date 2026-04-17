package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sage-roundtable/core/internal/llm"
	"gopkg.in/yaml.v3"
)

// LoadModelConfig 从 YAML 文件加载模型路由配置
func LoadModelConfig(filePath string) (*llm.RouterConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config llm.RouterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 处理环境变量替换
	for i := range config.Models {
		if err := expandEnvVars(&config.Models[i]); err != nil {
			return nil, fmt.Errorf("failed to expand env vars for model %s: %w", config.Models[i].Name, err)
		}
	}

	return &config, nil
}

// expandEnvVars 展开配置中的环境变量
// 支持 ${VAR_NAME} 格式
func expandEnvVars(model *llm.ModelConfig) error {
	model.APIKey = expandString(model.APIKey)
	model.SecretKey = expandString(model.SecretKey)
	model.BaseURL = expandString(model.BaseURL)
	model.Model = expandString(model.Model)
	model.Name = expandString(model.Name)

	// 展开元数据中的变量
	for k, v := range model.Metadata {
		model.Metadata[k] = expandString(v)
	}

	return nil
}

// expandString 展开单个字符串中的环境变量
func expandString(s string) string {
	// 匹配 ${VAR_NAME} 格式
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	return re.ReplaceAllStringFunc(s, func(match string) string {
		// 提取变量名
		varName := match[2 : len(match)-1] // 去掉 ${ 和 }

		// 获取环境变量值
		envValue := os.Getenv(varName)
		if envValue == "" {
			// 如果环境变量未设置，返回原始字符串（稍后可能会报错）
			return match
		}

		return envValue
	})
}

// ValidateModelConfig 验证模型配置的合法性
func ValidateModelConfig(config *llm.RouterConfig) error {
	if config.Strategy == "" {
		config.Strategy = llm.StrategyPriority
	}

	if len(config.Models) == 0 {
		return fmt.Errorf("no models configured")
	}

	enabledCount := 0
	for i, model := range config.Models {
		if model.Enabled {
			enabledCount++
		}

		// 验证必填字段
		if model.Name == "" {
			return fmt.Errorf("model[%d]: name is required", i)
		}

		if model.Type == "" {
			return fmt.Errorf("model[%d] (%s): type is required", i, model.Name)
		}

		if model.Model == "" {
			return fmt.Errorf("model[%d] (%s): model ID is required", i, model.Name)
		}

		// 根据类型验证特定字段
		switch model.Type {
		case llm.ModelTypeOpenAI, llm.ModelTypeQwen, llm.ModelTypeDeepSeek, llm.ModelTypeChatGLM, llm.ModelTypeOllama:
			if model.APIKey == "" && model.Type != llm.ModelTypeOllama {
				return fmt.Errorf("model[%d] (%s): api_key is required for %s", i, model.Name, model.Type)
			}

		case llm.ModelTypeWenxin:
			if model.APIKey == "" {
				return fmt.Errorf("model[%d] (%s): api_key is required for wenxin", i, model.Name)
			}
			if model.SecretKey == "" {
				return fmt.Errorf("model[%d] (%s): secret_key is required for wenxin", i, model.Name)
			}
		}
	}

	if enabledCount == 0 {
		return fmt.Errorf("at least one model must be enabled")
	}

	// 验证默认模型是否存在且已启用
	if config.DefaultModel != "" {
		found := false
		for _, model := range config.Models {
			if model.Name == config.DefaultModel {
				if !model.Enabled {
					return fmt.Errorf("default model '%s' is not enabled", config.DefaultModel)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("default model '%s' not found in models list", config.DefaultModel)
		}
	}

	return nil
}

// GetEnabledModels 获取所有已启用的模型配置
func GetEnabledModels(config *llm.RouterConfig) []llm.ModelConfig {
	var enabled []llm.ModelConfig
	for _, model := range config.Models {
		if model.Enabled {
			enabled = append(enabled, model)
		}
	}
	return enabled
}

// PrintModelSummary 打印模型配置摘要
func PrintModelSummary(config *llm.RouterConfig) {
	fmt.Println("\n📊 已配置的模型:")
	fmt.Println(strings.Repeat("-", 80))

	enabledCount := 0
	for _, model := range config.Models {
		status := "❌ 禁用"
		if model.Enabled {
			status = "✅ 启用"
			enabledCount++
		}

		fmt.Printf("%-20s | %-12s | %-20s | %s\n",
			model.Name,
			model.Type,
			model.Model,
			status,
		)
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("总计: %d 个模型, 已启用: %d 个\n\n", len(config.Models), enabledCount)
}
