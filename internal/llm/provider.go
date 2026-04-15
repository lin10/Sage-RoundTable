package llm

import "context"

// Message 表示对话消息
type Message struct {
	Role    string `json:"role"`    // system/user/assistant
	Content string `json:"content"`
	AgentID string `json:"agent_id,omitempty"` // 如果是智能体发言
}

// CompletionRequest 对话请求
type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	Stream      bool      `json:"stream"`
}

// CompletionResponse 对话响应
type CompletionResponse struct {
	ID      string     `json:"id"`
	Content string     `json:"content"`
	Usage   TokenUsage `json:"usage"`
}

// StreamChunk 流式响应块
type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

// TokenUsage Token使用统计
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Provider 定义了通用的 LLM 提供商接口
type Provider interface {
	// Complete 执行单次对话完成请求
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

	// StreamComplete 执行流式对话完成请求
	StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)

	// Name 返回提供商名称
	Name() string

	// Close 关闭连接，释放资源
	Close() error
}
