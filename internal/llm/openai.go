package llm

import (
	"context"
	"errors"
	"io"

	"github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client *openai.Client
	name   string
	model  string
}

// OpenAIConfig 包含了初始化 OpenAI Provider 所需的配置
type OpenAIConfig struct {
	Name    string // Provider 名称
	APIKey  string
	BaseURL string
	Model   string // 默认模型
}

// NewOpenAIProvider 创建基于 go-openai 的 Provider
// 同时也兼容支持任何兼容 OpenAI 接口规范的模型服务，比如通义千问 (Qwen) 或 DeepSeek
func NewOpenAIProvider(cfg OpenAIConfig) *OpenAIProvider {
	config := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}

	client := openai.NewClientWithConfig(config)
	return &OpenAIProvider{
		client: client,
		name:   cfg.Name,
		model:  cfg.Model,
	}
}

func (p *OpenAIProvider) Name() string {
	return p.name
}

func (p *OpenAIProvider) Close() error {
	return nil
}

func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// 如果请求中没有指定模型，使用配置的默认模型
	model := req.Model
	if model == "" {
		model = p.model
	}

	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: float32(req.Temperature),
			MaxTokens:   req.MaxTokens,
		},
	)

	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices returned from LLM")
	}

	return &CompletionResponse{
		ID:      resp.ID,
		Content: resp.Choices[0].Message.Content,
		Usage: TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (p *OpenAIProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	// 如果请求中没有指定模型，使用配置的默认模型
	model := req.Model
	if model == "" {
		model = p.model
	}

	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	stream, err := p.client.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: float32(req.Temperature),
			MaxTokens:   req.MaxTokens,
		},
	)

	if err != nil {
		return nil, err
	}

	chunks := make(chan StreamChunk)

	go func() {
		defer stream.Close()
		defer close(chunks)

		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				chunks <- StreamChunk{Done: true}
				return
			}
			if err != nil {
				// 对于流式错误，这里简单抛出一个带 error 的 chunk 或记录日志
				// 实际生产环境可以根据需求扩展 StreamChunk 结构包含 Error 字段
				return
			}

			if len(resp.Choices) > 0 {
				chunks <- StreamChunk{
					Content: resp.Choices[0].Delta.Content,
					Done:    false,
				}
			}
		}
	}()

	return chunks, nil
}
