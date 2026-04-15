package resolution

import (
	"context"
	"fmt"

	"github.com/sage-roundtable/core/internal/llm"
	"github.com/sage-roundtable/core/internal/session"
	"github.com/sage-roundtable/core/pkg/logger"
)

type Config struct {
	Model       string  `mapstructure:"model" yaml:"model"`
	Temperature float64 `mapstructure:"temperature" yaml:"temperature"`
}

type Generator struct {
	llmProvider llm.Provider
	config      *Config
}

func NewGenerator(provider llm.Provider, cfg *Config) *Generator {
	if cfg.Model == "" {
		cfg.Model = "gpt-4-turbo"
	}
	if cfg.Temperature == 0 {
		cfg.Temperature = 0.4 // 生成总结需要稳定和客观，温度偏低
	}
	return &Generator{
		llmProvider: provider,
		config:      cfg,
	}
}

// GenerateResolution 读取最终的会议白板，提炼并生成最终的《决策纪要》
func (g *Generator) GenerateResolution(ctx context.Context, cm *session.ContextManager) (string, error) {
	contextMD := cm.GetContent()

	prompt := fmt.Sprintf(`作为“赛博军机处”的首席书记官，请你基于以下所有先贤的辩论记录与会议白板，提炼并生成一份最终的《赛博军机处决策纪要》。

请严格按照以下 Markdown 格式输出（不要额外废话）：

# 📜 赛博军机处决策纪要

## 一、 决策背景
[一句话总结用户的核心诉求与关键约束]

## 二、 核心共识
[列出先贤们达成一致的 3-5 条核心观点，用 bullet points 呈现]

## 三、 主要分歧
[列出会议中出现的关键分歧点及代表人物，用 bullet points 呈现]

## 四、 上中下三策
### 🌟 上策（推荐方案：高收益高风险 或 最优解）
- **方案描述**：...
- **预期收益**：...
- **实施难度**：⭐⭐⭐（最多5星）
- **风险等级**：...

### ⚖️ 中策（稳健方案：折中、保守）
- **方案描述**：...
- **预期收益**：...
- **实施难度**：⭐⭐（最多5星）
- **风险等级**：...

### 🛡️ 下策（保底方案：托底、退路）
- **方案描述**：...
- **预期收益**：...
- **实施难度**：⭐（最多5星）
- **风险等级**：...

## 五、 关键风险提示
[罗列必须防范的黑天鹅事件或致命风险，尤其是魏征、马基雅维利等提出的风险]

## 六、 智者箴言
[挑选 1-2 位发言最精彩或最切中要害的先贤，附上他们的一句话总结]

---
【会议记录白板】：
%s`, contextMD)

	logger.Log.Info("Generating final resolution...")

	resp, err := g.llmProvider.Complete(ctx, llm.CompletionRequest{
		Model:       g.config.Model,
		Messages:    []llm.Message{{Role: "user", Content: prompt}},
		Temperature: g.config.Temperature,
		MaxTokens:   2500, // 决议书可能较长，放宽 Token 限制
	})

	if err != nil {
		logger.Log.Error("Failed to generate resolution", "error", err)
		return "", fmt.Errorf("failed to generate resolution: %w", err)
	}

	logger.Log.Info("Resolution generated successfully")
	return resp.Content, nil
}
