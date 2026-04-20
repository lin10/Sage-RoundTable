# 🏛️ Sage-RoundTable（赛博军机处）

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

> **人类从历史中吸取的最大教训是没有从历史中吸取教训。**

> **The only thing we learn from history is that we learn nothing from history.**

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

## 🏗️ 快速开始

### 前置条件
- Go 1.23+
- （可选）LLM API Key 或本地 Ollama

### 启动服务

项目提供两个独立的启动脚本：

#### 1️⃣ 后端 API 服务器

```bash
# Windows
.\start.bat

# Linux/Mac
chmod +x start.sh && ./start.sh
```

**功能：**
- 📡 REST API 服务
- 🗄️ 内存数据存储
- 🌐 监听端口: `8080`
- 🔗 健康检查: `http://localhost:8080/health`

#### 2️⃣ 前端 Web 服务器

```bash
# Windows
.\start-web.bat

# Linux/Mac
chmod +x start-web.sh && ./start-web.sh
```

**功能：**
- 🎨 测试界面
- 📱 响应式设计
- 🌐 监听端口: `3000`
- 🔗 访问地址: `http://localhost:3000/demo.html`

### 使用流程

1. **打开两个终端窗口**
2. **终端 1** - 运行 `start.bat` 启动后端
3. **终端 2** - 运行 `start-web.bat` 启动前端
4. **浏览器访问** `http://localhost:3000/demo.html`

### 配置模型（可选）

如果需要连接真实的 LLM，请配置模型：

```bash
# 复制配置文件
cp config/models.example.yaml config/models.yaml

# 编辑配置文件，填入你的 API Keys
# 支持：通义千问、智谱AI、文心一言、DeepSeek、Ollama 等
```

详细配置说明见 [多模型路由系统](docs/MODEL_ROUTING.md)

---

## 📡 API 接口使用

服务器启动后，可以通过 REST API 进行交互。

### 快速测试

#### 1. 健康检查
```bash
curl http://localhost:8080/health
```

#### 2. 创建会话
```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "initial_query": "是否应该投资人工智能领域？",
    "user_id": "test_user",
    "context_info": {
      "budget": "100万",
      "timeline": "1年"
    }
  }'
```

#### 3. 生成追问问题
```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/questions
```

#### 4. 提交答案
```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/answers \
  -H "Content-Type: application/json" \
  -d '{
    "answers": {
      "question_1": "预算约100万",
      "question_2": "期望1年内见效"
    }
  }'
```

#### 5. 开始辩论
```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/debate/start
```

#### 6. 生成决议书
```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/resolution/generate
```

### Postman 使用

可以导入以下 Collection 进行测试（将 `{session_id}` 替换为实际返回的 ID）：

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/v1/sessions` | 创建会话 |
| GET | `/api/v1/sessions/{id}` | 获取会话详情 |
| POST | `/api/v1/sessions/{id}/questions` | 生成问题 |
| POST | `/api/v1/sessions/{id}/answers` | 提交答案 |
| POST | `/api/v1/sessions/{id}/debate/start` | 开始辩论 |
| GET | `/api/v1/sessions/{id}/debate/stream` | 流式获取辩论（SSE） |
| POST | `/api/v1/sessions/{id}/resolution/generate` | 生成决议 |
| GET | `/api/v1/sessions/{id}/resolution` | 获取决议书 |
| GET | `/api/v1/config/agents` | 获取智者列表 |

完整的 API 文档请参考项目中的接口定义。

---

## 📚 项目结构

```
Sage-RoundTable/
├── cmd/
│   ├── cli/          # CLI 命令行工具
│   └── server/       # HTTP 服务器入口
├── internal/
│   ├── agent/        # 智能体管理
│   ├── api/          # API 路由和处理器
│   ├── config/       # 配置加载
│   ├── engine/       # 核心引擎（辩论、提问、决议）
│   ├── llm/          # LLM 模型路由
│   ├── session/      # 会话管理（内存/SQLite）
│   └── skill/        # 技能系统
├── pkg/
│   ├── cache/        # 缓存实现
│   ├── errors/       # 错误处理
│   ├── logger/       # 日志系统
│   └── utils/        # 工具函数
├── config/           # 配置文件
│   ├── agents/       # 智能体配置
│   └── skills/       # 技能配置
└── docs/             # 项目文档
```

---

## 🛠️ 开发说明

### 技术栈

- **后端框架**：原生 Go + chi 路由器
- **数据存储**：内存存储（MVP）/ SQLite（完整版）
- **LLM 集成**：支持多模型路由（OpenAI、通义千问、智谱AI、文心一言、DeepSeek、Ollama）
- **并发模型**：Go goroutines + channels
- **API 风格**：RESTful + Server-Sent Events (SSE)

### 添加新智能体

1. 在 `config/agents/` 目录下创建新的 `.md` 文件
2. 定义智能体的角色、思维模型和 Prompt 模板
3. 重启服务器即可使用

示例：`config/agents/new_sage.md`
```markdown
# 新智能体名称

**角色**: 角色描述
**专长**: 专业领域

## 思维模型

你的核心思考方式...

## 对话风格

你的语言特点...
```

---

## ❓ 常见问题

### Q: 为什么启动时报错 "CGO_ENABLED=0"？
A: MVP 版本已改用内存存储，不需要 CGO。直接运行 `go run cmd/server/main.go` 即可。

### Q: 如何切换到 SQLite 存储？
A: 安装 GCC 后，修改 `cmd/server/main.go`，将 `session.NewMemoryStore()` 改为 `session.NewSQLiteStore(dbPath)`。

### Q: 支持哪些 LLM 模型？
A: 支持 OpenAI GPT、阿里云通义千问、智谱AI GLM、百度文心一言、DeepSeek、ChatGLM 以及本地 Ollama 模型。

### Q: 数据会丢失吗？
A: MVP 版本使用内存存储，重启服务器后数据会丢失。如需持久化，请使用 SQLite 版本。

---

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

## 🌟 Star History

如果这个项目对你有帮助，请给我们一个 ⭐️！
