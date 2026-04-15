package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sage-roundtable/core/pkg/utils"
)

// Repository 负责管理所有的智能体配置
type Repository struct {
	agents map[string]*AgentProfile
}

func NewRepository() *Repository {
	return &Repository{
		agents: make(map[string]*AgentProfile),
	}
}

// LoadFromDir 从指定目录加载所有 .md 格式的智能体配置
func (r *Repository) LoadFromDir(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read agents directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		fullPath := filepath.Join(dirPath, entry.Name())
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read agent file %s: %w", fullPath, err)
		}

		profile := &AgentProfile{}
		body, err := utils.ParseMarkdownConfig(content, profile)
		if err != nil {
			return fmt.Errorf("failed to parse agent config from %s: %w", fullPath, err)
		}
		
		profile.SystemPrompt = body
		
		if profile.ID == "" {
			return fmt.Errorf("agent profile in %s is missing 'id'", fullPath)
		}

		r.agents[profile.ID] = profile
	}

	return nil
}

// Get 根据 ID 获取智能体配置
func (r *Repository) Get(id string) (*AgentProfile, bool) {
	agent, ok := r.agents[id]
	return agent, ok
}

// GetAll 获取所有已加载的智能体配置
func (r *Repository) GetAll() []*AgentProfile {
	var list []*AgentProfile
	for _, a := range r.agents {
		list = append(list, a)
	}
	return list
}

// --- DefaultAgent 默认实现 ---

type DefaultAgent struct {
	profile     *AgentProfile
	llmProvider llm.Provider
}

func NewDefaultAgent(profile *AgentProfile, provider llm.Provider) *DefaultAgent {
	return &DefaultAgent{
		profile:     profile,
		llmProvider: provider,
	}
}

func (a *DefaultAgent) ID() string {
	return a.profile.ID
}

func (a *DefaultAgent) Name() string {
	return a.profile.Name
}

func (a *DefaultAgent) Role() string {
	return a.profile.Role
}

func (a *DefaultAgent) Prompt() string {
	return a.profile.SystemPrompt
}

func (a *DefaultAgent) Metadata() map[string]interface{} {
	return map[string]interface{}{
		"tags":        a.profile.Tags,
		"temperature": a.profile.Temperature,
		"personality": a.profile.Personality,
	}
}

func (a *DefaultAgent) Think(ctx context.Context, contextMD string) (string, error) {
	// 构建该先贤专属的 Prompt：人设 + 当前会议白板
	prompt := fmt.Sprintf(`%s

【当前会议记录与讨论进展】：
%s

请基于你的思维特点和发言风格，对当前的议题发表你的见解，或者反驳其他人的观点。
直接输出你的发言内容，无需其他多余的客套话。`, a.profile.SystemPrompt, contextMD)

	resp, err := a.llmProvider.Complete(ctx, llm.CompletionRequest{
		Model:       "gpt-4-turbo", // 或者通过配置传递
		Messages:    []llm.Message{{Role: "user", Content: prompt}},
		Temperature: a.profile.Temperature,
		MaxTokens:   a.profile.MaxTokens,
	})

	if err != nil {
		return "", err
	}

	return resp.Content, nil
}
