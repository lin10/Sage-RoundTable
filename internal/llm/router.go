package llm

import (
	"context"
	"fmt"
	"sync"
)

// ModelType 定义模型类型
type ModelType string

const (
	ModelTypeOpenAI   ModelType = "openai"
	ModelTypeQwen     ModelType = "qwen"     // 阿里云通义千问
	ModelTypeWenxin   ModelType = "wenxin"   // 百度文心一言
	ModelTypeChatGLM  ModelType = "chatglm"  // 智谱AI ChatGLM
	ModelTypeDeepSeek ModelType = "deepseek" // DeepSeek
	ModelTypeOllama   ModelType = "ollama"   // 本地 Ollama
	ModelTypeClaude   ModelType = "claude"   // Anthropic Claude
	ModelTypeGemini   ModelType = "gemini"   // Google Gemini
)

// RoutingStrategy 定义路由策略
type RoutingStrategy string

const (
	StrategyFixed       RoutingStrategy = "fixed"       // 固定使用指定模型
	StrategyPriority    RoutingStrategy = "priority"    // 优先级顺序尝试
	StrategyLoadBalance RoutingStrategy = "loadbalance" // 负载均衡
	StrategyCostOptimal RoutingStrategy = "costoptimal" // 成本最优
)

// ModelConfig 单个模型的配置
type ModelConfig struct {
	Name         string            `yaml:"name" json:"name"`                     // 模型名称标识
	Type         ModelType         `yaml:"type" json:"type"`                     // 模型类型
	Model        string            `yaml:"model" json:"model"`                   // 具体模型ID (如 gpt-4, qwen-max)
	BaseURL      string            `yaml:"base_url" json:"base_url"`             // API Base URL
	APIKey       string            `yaml:"api_key" json:"api_key"`               // API Key
	SecretKey    string            `yaml:"secret_key" json:"secret_key"`         // 额外的密钥 (如文心一言需要)
	Temperature  float64           `yaml:"temperature" json:"temperature"`       // 默认温度
	MaxTokens    int               `yaml:"max_tokens" json:"max_tokens"`         // 最大Token数
	Priority     int               `yaml:"priority" json:"priority"`             // 优先级 (数字越小优先级越高)
	CostPerToken float64           `yaml:"cost_per_token" json:"cost_per_token"` // 每Token成本 (用于成本优化策略)
	Enabled      bool              `yaml:"enabled" json:"enabled"`               // 是否启用
	Metadata     map[string]string `yaml:"metadata" json:"metadata"`             // 额外元数据
}

// RouterConfig 路由器配置
type RouterConfig struct {
	DefaultModel string          `yaml:"default_model" json:"default_model"` // 默认模型名称
	Strategy     RoutingStrategy `yaml:"strategy" json:"strategy"`           // 路由策略
	Models       []ModelConfig   `yaml:"models" json:"models"`               // 所有可用模型
	Fallback     bool            `yaml:"fallback" json:"fallback"`           // 是否启用降级
	MaxRetries   int             `yaml:"max_retries" json:"max_retries"`     // 最大重试次数
}

// ModelRouter 模型路由器接口
type ModelRouter interface {
	// GetProvider 根据策略获取合适的 Provider
	GetProvider(ctx context.Context, purpose string) (Provider, error)

	// GetProviderByName 根据名称获取特定 Provider
	GetProviderByName(name string) (Provider, error)

	// GetAllProviders 获取所有可用的 Providers
	GetAllProviders() map[string]Provider

	// RegisterProvider 注册一个新的 Provider
	RegisterProvider(name string, provider Provider, config ModelConfig) error

	// Close 关闭所有 Providers
	Close() error
}

// router 模型路由器的实现
type router struct {
	mu           sync.RWMutex
	providers    map[string]Provider
	configs      map[string]ModelConfig
	routerConfig *RouterConfig
	defaultModel string
	strategy     RoutingStrategy
	fallback     bool
	maxRetries   int
	priorityList []string // 按优先级排序的模型列表
}

// NewRouter 创建新的模型路由器
func NewRouter(config *RouterConfig) (ModelRouter, error) {
	r := &router{
		providers:    make(map[string]Provider),
		configs:      make(map[string]ModelConfig),
		routerConfig: config,
		defaultModel: config.DefaultModel,
		strategy:     config.Strategy,
		fallback:     config.Fallback,
		maxRetries:   config.MaxRetries,
	}

	if r.maxRetries == 0 {
		r.maxRetries = 3
	}

	// 注册所有配置的模型
	for _, modelCfg := range config.Models {
		if !modelCfg.Enabled {
			continue
		}

		provider, err := createProvider(modelCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider for model %s: %w", modelCfg.Name, err)
		}

		r.providers[modelCfg.Name] = provider
		r.configs[modelCfg.Name] = modelCfg
	}

	// 构建优先级列表
	r.buildPriorityList()

	// 如果没有设置默认模型，使用优先级最高的
	if r.defaultModel == "" && len(r.priorityList) > 0 {
		r.defaultModel = r.priorityList[0]
	}

	return r, nil
}

// createProvider 根据配置创建对应的 Provider
func createProvider(cfg ModelConfig) (Provider, error) {
	switch cfg.Type {
	case ModelTypeOpenAI, ModelTypeDeepSeek:
		// DeepSeek 也使用 OpenAI 兼容接口
		return NewOpenAIProvider(OpenAIConfig{
			Name:    cfg.Name,
			APIKey:  cfg.APIKey,
			BaseURL: cfg.BaseURL,
			Model:   cfg.Model,
		}), nil

	case ModelTypeQwen:
		// 通义千问使用 OpenAI 兼容接口
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		}
		return NewOpenAIProvider(OpenAIConfig{
			Name:    cfg.Name,
			APIKey:  cfg.APIKey,
			BaseURL: baseURL,
			Model:   cfg.Model,
		}), nil

	case ModelTypeOllama:
		// Ollama 使用 OpenAI 兼容接口
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434/v1"
		}
		return NewOpenAIProvider(OpenAIConfig{
			Name:    cfg.Name,
			APIKey:  cfg.APIKey, // Ollama 通常不需要 API Key
			BaseURL: baseURL,
			Model:   cfg.Model,
		}), nil

	case ModelTypeWenxin:
		// 文心一言需要特殊处理
		return NewWenxinProvider(WenxinConfig{
			Name:      cfg.Name,
			APIKey:    cfg.APIKey,
			SecretKey: cfg.SecretKey,
			Model:     cfg.Model,
		})

	case ModelTypeChatGLM:
		// 智谱AI使用 OpenAI 兼容接口
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = "https://open.bigmodel.cn/api/paas/v4"
		}
		return NewOpenAIProvider(OpenAIConfig{
			Name:    cfg.Name,
			APIKey:  cfg.APIKey,
			BaseURL: baseURL,
			Model:   cfg.Model,
		}), nil

	case ModelTypeClaude:
		// TODO: 实现 Claude Provider
		return nil, fmt.Errorf("Claude provider not yet implemented")

	case ModelTypeGemini:
		// TODO: 实现 Gemini Provider
		return nil, fmt.Errorf("Gemini provider not yet implemented")

	default:
		return nil, fmt.Errorf("unsupported model type: %s", cfg.Type)
	}
}

// buildPriorityList 构建优先级列表
func (r *router) buildPriorityList() {
	type priorityItem struct {
		name     string
		priority int
	}

	var items []priorityItem
	for name, cfg := range r.configs {
		items = append(items, priorityItem{name: name, priority: cfg.Priority})
	}

	// 简单冒泡排序
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].priority > items[j].priority {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	r.priorityList = make([]string, len(items))
	for i, item := range items {
		r.priorityList[i] = item.name
	}
}

// GetProvider 根据策略获取合适的 Provider
func (r *router) GetProvider(ctx context.Context, purpose string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch r.strategy {
	case StrategyFixed:
		return r.getFixedProvider()

	case StrategyPriority:
		return r.getPriorityProvider()

	case StrategyLoadBalance:
		return r.getLoadBalancedProvider()

	case StrategyCostOptimal:
		return r.getCostOptimalProvider()

	default:
		return r.getFixedProvider()
	}
}

// getFixedProvider 获取固定模型
func (r *router) getFixedProvider() (Provider, error) {
	if r.defaultModel == "" {
		return nil, fmt.Errorf("no default model configured")
	}

	provider, ok := r.providers[r.defaultModel]
	if !ok {
		return nil, fmt.Errorf("default model %s not found", r.defaultModel)
	}

	return provider, nil
}

// getPriorityProvider 按优先级获取模型
func (r *router) getPriorityProvider() (Provider, error) {
	for _, name := range r.priorityList {
		if provider, ok := r.providers[name]; ok {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("no available providers")
}

// getLoadBalancedProvider 负载均衡获取模型 (简单轮询)
func (r *router) getLoadBalancedProvider() (Provider, error) {
	// TODO: 实现更复杂的负载均衡算法
	return r.getPriorityProvider()
}

// getCostOptimalProvider 成本最优获取模型
func (r *router) getCostOptimalProvider() (Provider, error) {
	// 按成本排序，选择最便宜的可用模型
	// TODO: 实现基于成本的排序
	return r.getPriorityProvider()
}

// GetProviderByName 根据名称获取特定 Provider
func (r *router) GetProviderByName(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// GetAllProviders 获取所有可用的 Providers
func (r *router) GetAllProviders() map[string]Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Provider)
	for name, provider := range r.providers {
		result[name] = provider
	}

	return result
}

// RegisterProvider 注册一个新的 Provider
func (r *router) RegisterProvider(name string, provider Provider, config ModelConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers[name] = provider
	r.configs[name] = config

	// 重新构建优先级列表
	r.buildPriorityList()

	return nil
}

// Close 关闭所有 Providers
func (r *router) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error
	for name, provider := range r.providers {
		if err := provider.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close provider %s: %w", name, err)
		}
	}

	return lastErr
}
