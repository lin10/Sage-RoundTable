package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sage-roundtable/core/internal/agent"
	cfg "github.com/sage-roundtable/core/internal/config"
	"github.com/sage-roundtable/core/internal/engine/debate"
	"github.com/sage-roundtable/core/internal/engine/questioning"
	"github.com/sage-roundtable/core/internal/engine/resolution"
	"github.com/sage-roundtable/core/internal/llm"
	"github.com/sage-roundtable/core/internal/session"
	"github.com/sage-roundtable/core/pkg/logger"
	"github.com/sage-roundtable/core/pkg/utils"
)

func main() {
	// 1. 初始化日志
	logger.Init("info", "text")
	fmt.Println("🚀 欢迎来到 Sage-RoundTable (赛博军机处) CLI 测试版")

	// 1.5 加载 .env 文件（如果存在）
	if err := utils.LoadDefaultEnvFile(); err != nil {
		fmt.Printf("⚠️  加载 .env 文件失败: %v\n", err)
	}

	// 2. 加载模型路由配置
	configPath := "./config/models.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("⚠️  配置文件 %s 不存在，请先复制 config/models.example.yaml 并配置\n", configPath)
		fmt.Println("💡 提示: cp config/models.example.yaml config/models.yaml")
		return
	}

	modelConfig, err := cfg.LoadModelConfig(configPath)
	if err != nil {
		fmt.Printf("❌ 加载模型配置失败: %v\n", err)
		return
	}

	// 验证配置
	if err := cfg.ValidateModelConfig(modelConfig); err != nil {
		fmt.Printf("❌ 模型配置验证失败: %v\n", err)
		return
	}

	// 打印模型摘要
	cfg.PrintModelSummary(modelConfig)

	// 创建模型路由器
	router, err := llm.NewRouter(modelConfig)
	if err != nil {
		fmt.Printf("❌ 创建模型路由器失败: %v\n", err)
		return
	}
	defer router.Close()

	// 获取默认 Provider
	llmProvider, err := router.GetProvider(context.Background(), "default")
	if err != nil {
		fmt.Printf("❌ 获取默认模型失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 使用模型: %s\n", modelConfig.DefaultModel)

	// 3. 加载 Agent
	agentRepo := agent.NewRepository()
	if err := agentRepo.LoadFromDir("./config/agents"); err != nil {
		fmt.Printf("❌ 无法加载 Agent 配置: %v\n", err)
		return
	}
	var agents []agent.Agent
	for _, p := range agentRepo.GetAll() {
		agents = append(agents, agent.NewDefaultAgent(p, llmProvider))
		fmt.Printf("✅ 加载先贤: %s (%s)\n", p.Name, p.Role)
	}

	if len(agents) == 0 {
		fmt.Println("❌ 没有找到任何先贤配置，请检查 config/agents 目录")
		return
	}

	// 4. 获取用户议题
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n💬 请输入您的决策困境 (例如: 我手里有 500 万，应该买房还是创业？): ")
	query, _ := reader.ReadString('\n')
	query = strings.TrimSpace(query)

	// 5. 初始化会话与白板
	cm := session.NewContextManager("test-session-1", query, llmProvider, 4000)

	// 6. 反向提问引擎
	fmt.Println("\n🔍 [阶段 1] 正在分析您的背景...")
	qEngine := questioning.NewQuestioningEngine(llmProvider, &questioning.Config{MinQuestions: 2, MaxQuestions: 3})
	questions, err := qEngine.GenerateQuestions(context.Background(), cm.GetContent())
	if err != nil {
		fmt.Printf("❌ 提问引擎失败: %v\n", err)
		return
	}

	answers := make(map[string]string)
	for i, q := range questions {
		fmt.Printf("\n❓ 追问 %d: %s\n", i+1, q)
		fmt.Print("👉 您的回答: ")
		ans, _ := reader.ReadString('\n')
		answers[q] = strings.TrimSpace(ans)
	}
	qEngine.HandleAnswers(cm, answers)

	// 7. 圆桌辩论引擎
	fmt.Println("\n🔥 [阶段 2] 先贤们正在激烈辩论中，请稍候...")
	dEngine := debate.NewDebateEngine(agents, llmProvider, &debate.Config{Rounds: 1, ParallelExecution: true})
	if err := dEngine.ExecuteDebate(context.Background(), cm); err != nil {
		fmt.Printf("❌ 辩论引擎失败: %v\n", err)
		return
	}

	// 打印一下内部白板的情况
	fmt.Println("\n================ [会议记录摘要] ================")
	fmt.Println(cm.GetContent())
	fmt.Println("================================================")

	// 8. 生成决议书
	fmt.Println("\n📜 [阶段 3] 正在生成《赛博军机处决策纪要》...")
	rGenerator := resolution.NewGenerator(llmProvider, &resolution.Config{})
	res, err := rGenerator.GenerateResolution(context.Background(), cm)
	if err != nil {
		fmt.Printf("❌ 生成决议书失败: %v\n", err)
		return
	}

	fmt.Println("\n" + res)
	fmt.Println("\n✅ 会议结束！")
}
