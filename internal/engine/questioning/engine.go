package questioning

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sage-roundtable/core/internal/llm"
	"github.com/sage-roundtable/core/internal/session"
	"github.com/sage-roundtable/core/pkg/logger"
)

type Config struct {
	MinQuestions int    `mapstructure:"min_questions" yaml:"min_questions"`
	MaxQuestions int    `mapstructure:"max_questions" yaml:"max_questions"`
	SkipEnabled  bool   `mapstructure:"skip_enabled" yaml:"skip_enabled"`
	Model        string `mapstructure:"model" yaml:"model"` // 使用的模型
}

type QuestioningEngine struct {
	llmProvider llm.Provider
	config      *Config
}

func NewQuestioningEngine(provider llm.Provider, cfg *Config) *QuestioningEngine {
	if cfg.Model == "" {
		cfg.Model = "gpt-4-turbo"
	}
	if cfg.MinQuestions == 0 {
		cfg.MinQuestions = 3
	}
	if cfg.MaxQuestions == 0 {
		cfg.MaxQuestions = 5
	}
	return &QuestioningEngine{
		llmProvider: provider,
		config:      cfg,
	}
}

// GenerateQuestions 核心功能：基于当前的 Markdown 上下文分析出还需要哪些关键信息
// 返回一个问题列表
func (e *QuestioningEngine) GenerateQuestions(ctx context.Context, contextMD string) ([]string, error) {
	// 1. 构建分析提示词
	prompt := e.buildAnalysisPrompt(contextMD)

	logger.Log.Debug("Requesting reverse questions from LLM", "prompt", prompt)

	// 2. 调用 LLM 识别信息缺口
	resp, err := e.llmProvider.Complete(ctx, llm.CompletionRequest{
		Model:       e.config.Model,
		Messages:    []llm.Message{{Role: "user", Content: prompt}},
		Temperature: 0.7, // 保持适中的创造力，提出好问题
		MaxTokens:   1000,
	})

	if err != nil {
		logger.Log.Error("Failed to generate questions via LLM", "error", err)
		return nil, fmt.Errorf("generate questions failed: %w", err)
	}

	// 3. 解析 LLM 返回的 JSON 数组
	var questions []string
	if err := json.Unmarshal([]byte(resp.Content), &questions); err != nil {
		// 容错处理：如果大模型没有严格返回 JSON 格式，可以尝试使用简单的正则或行切割。这里简化处理。
		logger.Log.Warn("Failed to unmarshal JSON questions, LLM raw response might be dirty", "content", resp.Content)
		return nil, fmt.Errorf("failed to parse LLM response to JSON array: %w", err)
	}

	// 4. 限制问题数量
	return e.limitQuestions(questions), nil
}

// buildAnalysisPrompt 构建反向提问的 Prompt
func (e *QuestioningEngine) buildAnalysisPrompt(contextMD string) string {
	return fmt.Sprintf(`你是一位专业的商业与人生决策分析师。
你的任务是：阅读下方用户的【初始议题与已知事实】，分析目前还缺少哪些关键信息才能给出全面的建议。
比如：用户的财务状况、隐性约束条件、时间周期、风险偏好等。

请针对决策的核心要素，提出需要追问的 %d 到 %d 个关键问题。
要求：
1. 问题必须具体且可回答，不要过于空泛。
2. 绝对不能重复已经提供的事实。
3. 请以严格的 JSON 数组格式返回问题列表（如 ["问题1", "问题2"]），不要包含其他多余文本。

【当前会议上下文白板】：
%s`, e.config.MinQuestions, e.config.MaxQuestions, contextMD)
}

// limitQuestions 根据配置限制最终问题数量
func (e *QuestioningEngine) limitQuestions(questions []string) []string {
	if len(questions) > e.config.MaxQuestions {
		return questions[:e.config.MaxQuestions]
	}
	return questions
}

// HandleAnswers 将用户对追问的回答更新回会议白板上下文
func (e *QuestioningEngine) HandleAnswers(cm *session.ContextManager, qas map[string]string) error {
	for q, a := range qas {
		fact := fmt.Sprintf("追问: %s \n回答: %s", q, a)
		if err := cm.AppendFact(fact); err != nil {
			return fmt.Errorf("failed to append fact: %w", err)
		}
	}
	return nil
}
