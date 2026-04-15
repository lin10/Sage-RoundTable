package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sage-roundtable/core/internal/llm"
	"github.com/sage-roundtable/core/pkg/logger"
	"github.com/sage-roundtable/core/pkg/utils"
)

// ContextManager 负责管理基于 Markdown 的会话白板
type ContextManager struct {
	mu         sync.RWMutex
	sessionID  string
	contentMD  string // 实时维护的 Markdown 上下文内容
	tokenCount int
	compressor *ContextCompressor
}

// ContextCompressor 上下文压缩模块
type ContextCompressor struct {
	llmProvider llm.Provider
	threshold   int // 触发压缩的 Token 阈值，如 4000
}

// NewContextManager 创建一个新的 Markdown 上下文管理器
func NewContextManager(sessionID string, initialQuery string, provider llm.Provider, threshold int) *ContextManager {
	initialMD := fmt.Sprintf("# 决策会议: %s\n\n## 初始议题\n%s\n\n## 核心背景与已知事实\n", time.Now().Format("2006-01-02"), initialQuery)
	
	return &ContextManager{
		sessionID:  sessionID,
		contentMD:  initialMD,
		tokenCount: utils.EstimateTokens(initialMD),
		compressor: &ContextCompressor{
			llmProvider: provider,
			threshold:   threshold,
		},
	}
}

// AppendFact 追加客观事实或背景信息（例如用户回答了反向提问）
func (cm *ContextManager) AppendFact(fact string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	newFact := fmt.Sprintf("- %s\n", fact)
	cm.contentMD += newFact
	cm.tokenCount = utils.EstimateTokens(cm.contentMD)
	
	logger.Log.Debug("Fact appended to Context", "fact", fact, "tokens", cm.tokenCount)
	return cm.checkAndCompress()
}

// AppendStatement 将先贤的发言追加到 Markdown 上下文中
func (cm *ContextManager) AppendStatement(agentName, statement string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.contentMD += fmt.Sprintf("\n### %s 发言:\n%s\n", agentName, statement)
	cm.tokenCount = utils.EstimateTokens(cm.contentMD)
	
	logger.Log.Debug("Statement appended to Context", "agent", agentName, "tokens", cm.tokenCount)
	return cm.checkAndCompress()
}

// GetContent 获取当前的完整 Markdown 上下文
func (cm *ContextManager) GetContent() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.contentMD
}

// checkAndCompress 检查 Token 是否溢出并执行压缩
// 注意：此方法应在持有 cm.mu.Lock() 锁的情况下调用
func (cm *ContextManager) checkAndCompress() error {
	if cm.tokenCount <= cm.compressor.threshold {
		return nil
	}

	logger.Log.Info("Context token threshold reached, triggering compression", 
		"session", cm.sessionID, 
		"current_tokens", cm.tokenCount,
		"threshold", cm.compressor.threshold,
	)
	
	return cm.compressor.Compress(cm)
}

// Compress 执行大模型自动压缩机制
func (cc *ContextCompressor) Compress(cm *ContextManager) error {
	// 构建压缩 Prompt
	prompt := fmt.Sprintf(`请将以下会议记录压缩为精简的摘要。
要求：
1. 必须保留用户的核心议题与已知事实（非常重要！不要丢失数字和关键背景）。
2. 保留当前各位先贤（Agent）的核心观点、关键分歧与共识。
3. 丢弃无效的寒暄、语气词（如“臣以为”、“主公”）以及过度重复的修辞。
4. 保持清晰的 Markdown 结构输出，结构建议包含：“初始议题”、“核心事实”、“阶段性共识与分歧”。

以下是原始会议记录：
%s`, cm.contentMD)

	// 调用 LLM
	resp, err := cc.llmProvider.Complete(context.Background(), llm.CompletionRequest{
		Model:       "gpt-4-turbo", // 可配置为环境变量
		Messages:    []llm.Message{{Role: "user", Content: prompt}},
		Temperature: 0.3, // 低温度保证摘要的客观与稳定性
		MaxTokens:   2000,
	})

	if err != nil {
		logger.Log.Error("Context compression failed", "error", err)
		return fmt.Errorf("context compression failed: %w", err)
	}
	
	// 覆写原上下文
	cm.contentMD = "# 决策会议 (已摘要压缩)\n\n" + resp.Content + "\n\n## 后续讨论\n"
	cm.tokenCount = utils.EstimateTokens(cm.contentMD)
	
	logger.Log.Info("Context compression successful", 
		"session", cm.sessionID, 
		"new_tokens", cm.tokenCount,
	)

	return nil
}
