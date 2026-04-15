package debate

import (
	"context"
	"fmt"
	"sync"

	"github.com/sage-roundtable/core/internal/agent"
	"github.com/sage-roundtable/core/internal/llm"
	"github.com/sage-roundtable/core/internal/session"
	"github.com/sage-roundtable/core/pkg/logger"
)

type Config struct {
	Rounds            int  `mapstructure:"rounds" yaml:"rounds"`
	ParallelExecution bool `mapstructure:"parallel_execution" yaml:"parallel_execution"`
	Model             string
}

type DebateEngine struct {
	agents      []agent.Agent
	llmProvider llm.Provider
	moderator   *Moderator
	config      *Config
}

func NewDebateEngine(agents []agent.Agent, provider llm.Provider, cfg *Config) *DebateEngine {
	if cfg.Rounds <= 0 {
		cfg.Rounds = 2
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4-turbo"
	}
	return &DebateEngine{
		agents:      agents,
		llmProvider: provider,
		moderator:   NewModerator(provider, cfg.Model),
		config:      cfg,
	}
}

// ExecuteDebate 启动圆桌辩论流程，包含“独立发言”和“交叉质询”
func (e *DebateEngine) ExecuteDebate(ctx context.Context, cm *session.ContextManager) error {
	logger.Log.Info("Starting Debate Engine", "rounds", e.config.Rounds, "agents", len(e.agents))

	// 阶段一：先贤们并发或顺序进行首轮独立发言
	if err := e.executeIndividualStatements(ctx, cm); err != nil {
		return fmt.Errorf("individual statements failed: %w", err)
	}

	// 阶段二：多轮交叉质询（辩论）
	for i := 0; i < e.config.Rounds; i++ {
		logger.Log.Info("Starting Debate Round", "round", i+1)
		if err := e.executeCrossExamination(ctx, cm, i+1); err != nil {
			return fmt.Errorf("cross examination round %d failed: %w", i+1, err)
		}
	}

	logger.Log.Info("Debate Engine finished successfully")
	return nil
}

// executeIndividualStatements 第一阶段：每位先贤基于背景给出自己的独立见解
func (e *DebateEngine) executeIndividualStatements(ctx context.Context, cm *session.ContextManager) error {
	if e.config.ParallelExecution {
		return e.executeParallel(ctx, cm)
	}
	return e.executeSequential(ctx, cm)
}

// executeSequential 顺序执行，先贤依次发言，后面的能看到前面的发言
func (e *DebateEngine) executeSequential(ctx context.Context, cm *session.ContextManager) error {
	for _, a := range e.agents {
		contextMD := cm.GetContent()

		resp, err := a.Think(ctx, contextMD)
		if err != nil {
			logger.Log.Error("Agent think failed", "agent", a.Name(), "error", err)
			continue
		}

		if err := cm.AppendStatement(a.Name(), resp); err != nil {
			return err
		}
	}
	return nil
}

// executeParallel 并发执行，所有人同时针对同一个初始白板发言
func (e *DebateEngine) executeParallel(ctx context.Context, cm *session.ContextManager) error {
	contextMD := cm.GetContent()

	var wg sync.WaitGroup
	type result struct {
		agentName string
		content   string
		err       error
	}

	results := make(chan result, len(e.agents))

	for _, a := range e.agents {
		wg.Add(1)
		go func(ag agent.Agent) {
			defer wg.Done()
			resp, err := ag.Think(ctx, contextMD)
			results <- result{agentName: ag.Name(), content: resp, err: err}
		}(a)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集并写入上下文
	for r := range results {
		if r.err != nil {
			logger.Log.Error("Agent parallel think failed", "agent", r.agentName, "error", r.err)
			continue
		}
		if err := cm.AppendStatement(r.agentName, r.content); err != nil {
			return err
		}
	}

	return nil
}

// executeCrossExamination 第二阶段：主持人抓取分歧点，触发反驳与辩论
func (e *DebateEngine) executeCrossExamination(ctx context.Context, cm *session.ContextManager, roundNum int) error {
	// 1. 主持人识别冲突
	conflicts, err := e.moderator.IdentifyConflicts(ctx, cm)
	if err != nil {
		logger.Log.Warn("Moderator failed to identify conflicts, skipping this round", "error", err)
		return nil
	}

	if len(conflicts) == 0 {
		logger.Log.Info("Moderator found no major conflicts, debate can conclude early.")
		// 如果没有冲突，可以考虑提前结束（此处选择跳过该轮，也可选择 break 外层循环）
		return nil
	}

	// 2. 针对每个冲突点触发局部辩论
	for _, conflict := range conflicts {
		logger.Log.Info("Debating conflict", "topic", conflict.Topic, "participants", conflict.Participants)

		// 记录主持人的打断与引导
		moderatorPrompt := e.moderator.GenerateDebatePrompt(conflict)
		_ = cm.AppendStatement("主持人", moderatorPrompt)

		// 挑选相关的先贤参与辩论
		participants := e.selectDebateParticipants(conflict.Participants)
		if len(participants) == 0 {
			continue
		}

		// 辩论回合默认顺序发言，确保他们能相互反驳
		for _, participant := range participants {
			contextMD := cm.GetContent()

			resp, err := participant.Think(ctx, contextMD)
			if err != nil {
				logger.Log.Error("Debating participant failed to think", "agent", participant.Name(), "error", err)
				continue
			}

			if err := cm.AppendStatement(participant.Name(), resp); err != nil {
				return err
			}
		}
	}

	return nil
}

// selectDebateParticipants 根据冲突报告中的名单，从当前会议的先贤池中捞出相关先贤实例
func (e *DebateEngine) selectDebateParticipants(participantNames []string) []agent.Agent {
	var selected []agent.Agent
	
	// 建立 map 加快查找
	nameMap := make(map[string]bool)
	for _, name := range participantNames {
		nameMap[name] = true
	}

	for _, a := range e.agents {
		if nameMap[a.Name()] || nameMap[a.ID()] { // 兼容按名字或 ID 匹配
			selected = append(selected, a)
		}
	}
	return selected
}
