package agent

import "context"

// Agent 定义了先贤智能体的通用接口
type Agent interface {
	// ID 返回智能体的唯一标识
	ID() string

	// Name 返回智能体的显示名称
	Name() string

	// Role 返回智能体的角色定位
	Role() string

	// Think 基于上下文白板生成发言或思考结果
	Think(ctx context.Context, contextMD string) (string, error)

	// Prompt 获取该智能体的原始 Prompt 设定
	Prompt() string

	// Metadata 返回其他属性
	Metadata() map[string]interface{}
}

// AgentProfile 基于 Markdown Frontmatter 解析出的智能体配置
type AgentProfile struct {
	ID          string            `mapstructure:"id" yaml:"id" json:"id"`
	Name        string            `mapstructure:"name" yaml:"name" json:"name"`
	Role        string            `mapstructure:"role" yaml:"role" json:"role"`
	Description string            `mapstructure:"description" yaml:"description" json:"description"`
	Temperature float64           `mapstructure:"temperature" yaml:"temperature" json:"temperature"`
	MaxTokens   int               `mapstructure:"max_tokens" yaml:"max_tokens" json:"max_tokens"`
	Personality map[string]string `mapstructure:"personality" yaml:"personality" json:"personality"`
	Tags        []string          `mapstructure:"tags" yaml:"tags" json:"tags"`
	// 解析时将 Markdown 正文注入此字段
	SystemPrompt string `json:"system_prompt,omitempty"`
}
