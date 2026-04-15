# 🏛️ Sage-RoundTable（赛博军机处）

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

> 让人类历史上最杰出的 11 位智者，为你的每一个重大决策召开一场“圆桌会议”。

## ✨ 项目简介

**Sage-RoundTable** 是一个开源的多智能体（Multi-Agent）决策辅助工具。

你是否曾在买房、跳槽、创业或选专业时感到迷茫？个人的视角总有盲区。本项目利用大语言模型（LLM）扮演孔子、诸葛亮、亚当·斯密、马基雅维利等东西方先贤，通过 **“反向提问收集背景 → 圆桌辩论 → 生成决议书”** 的流程，为你提供跨越时空与学科的深度策略分析。

## 🎭 智囊团阵容（首批 11 人，6 东 5 西）

| 先贤 | 角色定位 | 一句话思维模型 |
| :--- | :--- | :--- |
| **诸葛亮** | 首席战略官 | 提出“上中下三策”，权衡天时地利人和 |
| **魏征** | 首席风险官 | 直言进谏，专攻逻辑漏洞与潜在隐患 |
| **管仲** | 经济顾问 | 核算成本收益，关注财政可持续性 |
| **张良** | 人心分析师 | 分析利益相关者，以柔克刚 |
| **司马懿** | 时机分析师 | 隐忍待时，评估对手弱点 |
| **孙子** | 博弈顾问 | 提炼博弈模型，追求“不战而屈人之兵” |
| **老子** | 哲学远见者 | 道法自然，提供超越功利的长期视角 |
| **亚里士多德** | 逻辑学导师 | 检验前提真伪，运用第一性原理 |
| **亚当·斯密** | 市场经济学家 | 分析“看不见的手”与长期供需关系 |
| **孟德斯鸠** | 法学顾问 | 审查合规性与权力边界 |
| **马基雅维利** | 现实政治顾问 | 抛开道德，推演最冷酷的现实博弈 |

## 🚀 核心特性

- **🤔 反向提问机制**：严禁直接给答案。系统必须先追问 3-5 个关键事实，补全决策背景。
- **🗣️ 圆桌会议引擎**：调度多位先贤并发发言，支持交叉质询与辩论。
- **📜 结构化决议书**：输出包含共识、分歧、上中下三策的《赛博军机处决策纪要》。
- **🧩 极简扩展**：新增角色只需添加一个 `.yaml` 配置文件和 Prompt 模板。
- **💸 成本友好**：支持本地 Ollama 模型，可在完全离线环境下运行。

## 📦 技术栈（原生 Go 实现）

本项目拒绝框架臃肿，基于 Go 标准库与极简依赖构建：

- **语言**：Go 1.23+（利用原生并发与 `net/http`）
- **LLM 接入**：多模型路由引擎（支持 OpenAI / Qwen / Wenxin / ChatGLM / DeepSeek / Ollama）
- **配置**：YAML + 环境变量
- **日志**：`slog`（标准库）
- **前端**：Next.js（React）或 Streamlit

## 🏗️ 快速开始（本地运行）

### 前置条件
- Go 1.23+
- 可访问的 LLM API（推荐阿里云通义千问、智谱AI等国内服务）或本地运行的 Ollama

### 步骤
```bash
# 1. 克隆仓库
git clone https://github.com/yourname/pantheon-council.git
cd pantheon-council

# 2. 配置模型路由（支持国内外多个大模型）
cp config/models.example.yaml config/models.yaml
# 编辑 config/models.yaml，启用你需要的模型

# 3. 设置 API Keys（任选其一）
# 方式一：使用 .env 文件
cp .env.example .env
# 编辑 .env 填入你的 API Keys

# 方式二：直接设置环境变量
export QWEN_API_KEY="sk-your-key"

# 4. 运行 CLI 测试版
go run cmd/cli/main.go
```

### 💡 推荐配置（国内用户）

```yaml
# config/models.yaml
default_model: "qwen-max"
strategy: "priority"

models:
  # 主力模型 - 阿里云通义千问
  - name: "qwen-max"
    type: "qwen"
    model: "qwen-max"
    api_key: "${QWEN_API_KEY}"
    enabled: true
    priority: 1
    
  # 备用模型 - 智谱AI
  - name: "glm-4"
    type: "chatglm"
    model: "glm-4"
    api_key: "${ZHIPU_API_KEY}"
    enabled: true
    priority: 2
    
  # 本地保底 - Ollama (完全免费)
  - name: "ollama-llama3"
    type: "ollama"
    model: "llama3"
    base_url: "http://localhost:11434/v1"
    enabled: true
    priority: 3
```

更多配置详情请查看 [多模型路由系统使用指南](docs/MODEL_ROUTING.md)
