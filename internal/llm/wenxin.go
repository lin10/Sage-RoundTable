package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WenxinConfig 文心一言配置
type WenxinConfig struct {
	Name      string
	APIKey    string
	SecretKey string
	Model     string // ernie-bot, ernie-bot-turbo, ernie-bot-4 等
}

// WenxinProvider 百度文心一言 Provider
type WenxinProvider struct {
	name      string
	apiKey    string
	secretKey string
	model     string
	accessToken string
	tokenExpireAt time.Time
	httpClient *http.Client
}

// wenxinTokenResponse 获取 Access Token 的响应
type wenxinTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Error       string `json:"error,omitempty"`
}

// wenxinMessage 文心一言消息格式
type wenxinMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// wenxinRequest 文心一言请求格式
type wenxinRequest struct {
	Messages    []wenxinMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	PenaltyScore float64        `json:"penalty_score,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// wenxinResponse 文心一言响应格式
type wenxinResponse struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Result  string          `json:"result"`
	IsEnd   bool            `json:"is_end"`
	Usage   wenxinTokenUsage `json:"usage"`
	ErrorCode int           `json:"error_code,omitempty"`
	ErrorMsg string         `json:"error_msg,omitempty"`
}

// wenxinTokenUsage Token使用统计
type wenxinTokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewWenxinProvider 创建文心一言 Provider
func NewWenxinProvider(cfg WenxinConfig) (*WenxinProvider, error) {
	p := &WenxinProvider{
		name:       cfg.Name,
		apiKey:     cfg.APIKey,
		secretKey:  cfg.SecretKey,
		model:      cfg.Model,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}

	// 初始化时获取 Access Token
	if err := p.refreshAccessToken(); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	return p, nil
}

// refreshAccessToken 刷新 Access Token
func (p *WenxinProvider) refreshAccessToken() error {
	url := fmt.Sprintf(
		"https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%s&client_secret=%s",
		p.apiKey,
		p.secretKey,
	)

	resp, err := p.httpClient.Post(url, "", nil)
	if err != nil {
		return fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token response: %w", err)
	}

	var tokenResp wenxinTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Error != "" {
		return fmt.Errorf("token error: %s", tokenResp.Error)
	}

	p.accessToken = tokenResp.AccessToken
	p.tokenExpireAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// ensureValidToken 确保 Access Token 有效
func (p *WenxinProvider) ensureValidToken() error {
	if p.accessToken == "" || time.Now().After(p.tokenExpireAt.Add(-5*time.Minute)) {
		return p.refreshAccessToken()
	}
	return nil
}

// getAPIURL 获取 API URL
func (p *WenxinProvider) getAPIURL() string {
	// 根据不同模型返回不同的 endpoint
	switch p.model {
	case "ernie-bot-4":
		return fmt.Sprintf(
			"https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions_pro?access_token=%s",
			p.accessToken,
		)
	case "ernie-bot-turbo":
		return fmt.Sprintf(
			"https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/eb-instant?access_token=%s",
			p.accessToken,
		)
	default: // ernie-bot
		return fmt.Sprintf(
			"https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions?access_token=%s",
			p.accessToken,
		)
	}
}

func (p *WenxinProvider) Name() string {
	return p.name
}

func (p *WenxinProvider) Close() error {
	return nil
}

func (p *WenxinProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// 确保 Token 有效
	if err := p.ensureValidToken(); err != nil {
		return nil, err
	}

	// 转换消息格式
	messages := make([]wenxinMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = wenxinMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 构建请求
	wenxinReq := wenxinRequest{
		Messages:    messages,
		Temperature: req.Temperature,
		TopP:        0.8,
		Stream:      false,
	}

	reqBody, err := json.Marshal(wenxinReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求
	apiURL := p.getAPIURL()
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var wenxinResp wenxinResponse
	if err := json.Unmarshal(body, &wenxinResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if wenxinResp.ErrorCode != 0 {
		return nil, fmt.Errorf("wenxin API error [%d]: %s", wenxinResp.ErrorCode, wenxinResp.ErrorMsg)
	}

	return &CompletionResponse{
		ID:      wenxinResp.ID,
		Content: wenxinResp.Result,
		Usage: TokenUsage{
			PromptTokens:     wenxinResp.Usage.PromptTokens,
			CompletionTokens: wenxinResp.Usage.CompletionTokens,
			TotalTokens:      wenxinResp.Usage.TotalTokens,
		},
	}, nil
}

func (p *WenxinProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	// TODO: 实现流式调用
	return nil, errors.New("streaming not yet implemented for Wenxin")
}
