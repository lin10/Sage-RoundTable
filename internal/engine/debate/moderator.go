package debate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sage-roundtable/core/internal/llm"
	"github.com/sage-roundtable/core/internal/session"
	"github.com/sage-roundtable/core/pkg/logger"
)

// Moderator 负责辅助辩论引擎，识别冲突并生成引导话术
type Moderator struct {
	llmProvider llm.Provider
	model       string
}

func NewModerator(provider llm.Provider, model string) *Moderator {
	if model == "" {
		model = "gpt-4-turbo"
	}
	return &Moderator{
		llmProvider: provider,
		model:       model,
	}
}

// Conflict 表示先贤之间的一个具体观点冲突
type Conflict struct {
	Topic        string   `json:"topic"`
	Participants []string `json:"participants"`
	Description  string   `json:"description"`
}

// IdentifyConflicts 基于当前的会议记录，识别出核心冲突点
func (m *Moderator) IdentifyConflicts(ctx context.Context, cm *session.ContextManager) ([]Conflict, error) {
	contextMD := cm.GetContent()
	prompt := fmt.Sprintf(`作为圆桌会议的主持人，请阅读以下会议记录，找出先贤们的核心观点分歧。
最多提取 3 个最重要、最值得深入辩论的冲突点。
必须严格以 JSON 数组格式返回，格式如下：
[
  {
    "topic": "分歧的主题/焦点",
    "participants": ["先贤A的名字", "先贤B的名字"],
    "description": "一句话描述双方的分歧所在"
  }
]

【会议记录】：
%s`, contextMD)

	logger.Log.Debug("Moderator is identifying conflicts")

	resp, err := m.llmProvider.Complete(ctx, llm.CompletionRequest{
		Model:       m.model,
		Messages:    []llm.Message{{Role: "user", Content: prompt}},
		Temperature: 0.2, // 提取事实，需要低温度
		MaxTokens:   1000,
	})

	if err != nil {
		return nil, fmt.Errorf("moderator failed to identify conflicts: %w", err)
	}

	var conflicts []Conflict
	if err := json.Unmarshal([]byte(resp.Content), &conflicts); err != nil {
		logger.Log.Warn("Moderator JSON parsing failed, returning empty conflicts", "err", err, "raw", resp.Content)
		return []Conflict{}, nil
	}

	return conflicts, nil
}

// GenerateDebatePrompt 根据冲突点，生成引导相关先贤进行反驳的提示词
func (m *Moderator) GenerateDebatePrompt(conflict Conflict) string {
	return fmt.Sprintf(`主持人打断：
各位先贤，我们注意到在【%s】这个问题上，%v 存在明显分歧。
分歧点在于：%s

请相关先贤针对这一分歧，直接反驳对方的逻辑漏洞或补充自己的论据。要求一针见血，直击要害！`, conflict.Topic, conflict.Participants, conflict.Description)
}
