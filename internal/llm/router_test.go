package llm

import (
	"context"
	"testing"
)

// TestRouterCreation 测试路由器创建
func TestRouterCreation(t *testing.T) {
	config := &RouterConfig{
		DefaultModel: "test-model",
		Strategy:     StrategyPriority,
		Fallback:     true,
		MaxRetries:   3,
		Models: []ModelConfig{
			{
				Name:     "test-model",
				Type:     ModelTypeOllama,
				Model:    "llama3",
				BaseURL:  "http://localhost:11434/v1",
				Enabled:  true,
				Priority: 1,
			},
		},
	}

	router, err := NewRouter(config)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}
	defer router.Close()

	if router == nil {
		t.Fatal("Router should not be nil")
	}
}

// TestGetProviderByName 测试根据名称获取Provider
func TestGetProviderByName(t *testing.T) {
	config := &RouterConfig{
		DefaultModel: "test-ollama",
		Strategy:     StrategyFixed,
		Models: []ModelConfig{
			{
				Name:     "test-ollama",
				Type:     ModelTypeOllama,
				Model:    "llama3",
				BaseURL:  "http://localhost:11434/v1",
				Enabled:  true,
				Priority: 1,
			},
		},
	}

	router, err := NewRouter(config)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}
	defer router.Close()

	provider, err := router.GetProviderByName("test-ollama")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	if provider.Name() != "test-ollama" {
		t.Errorf("Expected provider name 'test-ollama', got '%s'", provider.Name())
	}
}

// TestGetAllProviders 测试获取所有Providers
func TestGetAllProviders(t *testing.T) {
	config := &RouterConfig{
		DefaultModel: "model1",
		Strategy:     StrategyPriority,
		Models: []ModelConfig{
			{
				Name:     "model1",
				Type:     ModelTypeOllama,
				Model:    "llama3",
				BaseURL:  "http://localhost:11434/v1",
				Enabled:  true,
				Priority: 1,
			},
			{
				Name:     "model2",
				Type:     ModelTypeOllama,
				Model:    "qwen:7b",
				BaseURL:  "http://localhost:11434/v1",
				Enabled:  true,
				Priority: 2,
			},
		},
	}

	router, err := NewRouter(config)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}
	defer router.Close()

	providers := router.GetAllProviders()
	if len(providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(providers))
	}
}

// TestPriorityStrategy 测试优先级策略
func TestPriorityStrategy(t *testing.T) {
	config := &RouterConfig{
		DefaultModel: "",
		Strategy:     StrategyPriority,
		Models: []ModelConfig{
			{
				Name:     "high-priority",
				Type:     ModelTypeOllama,
				Model:    "llama3",
				BaseURL:  "http://localhost:11434/v1",
				Enabled:  true,
				Priority: 1,
			},
			{
				Name:     "low-priority",
				Type:     ModelTypeOllama,
				Model:    "qwen:7b",
				BaseURL:  "http://localhost:11434/v1",
				Enabled:  true,
				Priority: 10,
			},
		},
	}

	router, err := NewRouter(config)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}
	defer router.Close()

	provider, err := router.GetProvider(context.Background(), "default")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}

	// 应该返回优先级最高的模型
	if provider.Name() != "high-priority" {
		t.Errorf("Expected 'high-priority' provider, got '%s'", provider.Name())
	}
}

// TestDisabledModels 测试禁用的模型不会被加载
func TestDisabledModels(t *testing.T) {
	config := &RouterConfig{
		DefaultModel: "enabled-model",
		Strategy:     StrategyFixed,
		Models: []ModelConfig{
			{
				Name:     "enabled-model",
				Type:     ModelTypeOllama,
				Model:    "llama3",
				BaseURL:  "http://localhost:11434/v1",
				Enabled:  true,
				Priority: 1,
			},
			{
				Name:     "disabled-model",
				Type:     ModelTypeOllama,
				Model:    "qwen:7b",
				BaseURL:  "http://localhost:11434/v1",
				Enabled:  false,
				Priority: 2,
			},
		},
	}

	router, err := NewRouter(config)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}
	defer router.Close()

	providers := router.GetAllProviders()
	if len(providers) != 1 {
		t.Errorf("Expected 1 provider (disabled should be excluded), got %d", len(providers))
	}

	// 禁用的模型不应该存在
	if _, exists := providers["disabled-model"]; exists {
		t.Error("Disabled model should not be in providers")
	}
}

// TestOpenAICompatibleProvider 测试OpenAI兼容的Provider创建
func TestOpenAICompatibleProvider(t *testing.T) {
	tests := []struct {
		name      string
		modelType ModelType
		baseURL   string
	}{
		{"OpenAI", ModelTypeOpenAI, ""},
		{"Qwen", ModelTypeQwen, "https://dashscope.aliyuncs.com/compatible-mode/v1"},
		{"DeepSeek", ModelTypeDeepSeek, "https://api.deepseek.com/v1"},
		{"ChatGLM", ModelTypeChatGLM, "https://open.bigmodel.cn/api/paas/v4"},
		{"Ollama", ModelTypeOllama, "http://localhost:11434/v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &RouterConfig{
				DefaultModel: "test-model",
				Strategy:     StrategyFixed,
				Models: []ModelConfig{
					{
						Name:     "test-model",
						Type:     tt.modelType,
						Model:    "test",
						APIKey:   "test-key",
						BaseURL:  tt.baseURL,
						Enabled:  true,
						Priority: 1,
					},
				},
			}

			router, err := NewRouter(config)
			if err != nil {
				t.Fatalf("Failed to create router for %s: %v", tt.name, err)
			}
			defer router.Close()

			provider, err := router.GetProviderByName("test-model")
			if err != nil {
				t.Fatalf("Failed to get provider for %s: %v", tt.name, err)
			}

			if provider == nil {
				t.Errorf("Provider for %s should not be nil", tt.name)
			}
		})
	}
}
