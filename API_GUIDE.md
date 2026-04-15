# Sage-RoundTable API 服务器

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置环境变量

复制 `.env.example` 为 `.env` 并配置你的参数：

```bash
cp .env.example .env
```

编辑 `.env` 文件，设置必要的配置：

```env
# 服务器配置
SERVER_PORT=8080
DB_PATH=./data/sage_roundtable.db
CACHE_TTL=3600
RATE_LIMIT=60

# LLM API Keys (根据需要配置)
OPENAI_API_KEY=sk-your-key
QWEN_API_KEY=sk-your-key
```

### 3. 创建数据目录

```bash
mkdir -p data
```

### 4. 启动服务器

```bash
go run cmd/server/main.go
```

服务器将在 `http://localhost:8080` 启动。

## API 端点

### 健康检查

```bash
curl http://localhost:8080/health
```

响应：
```json
{
  "status": "healthy"
}
```

### 会话管理

#### 创建会话

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "initial_query": "我应该辞职创业吗？",
    "user_id": "user123"
  }'
```

响应：
```json
{
  "session_id": "sess_我应该辞...",
  "state": "initialized"
}
```

#### 获取会话详情

```bash
curl http://localhost:8080/api/v1/sessions/{session_id}
```

#### 列出会话

```bash
curl http://localhost:8080/api/v1/sessions?user_id=user123
```

#### 删除会话

```bash
curl -X DELETE http://localhost:8080/api/v1/sessions/{session_id}
```

### 提问阶段

#### 生成追问问题

```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/questions
```

响应：
```json
{
  "questions": [
    "您的预算范围是多少？",
    "您期望的时间周期是多久？",
    "您的风险承受能力如何？"
  ]
}
```

#### 提交答案

```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/answers \
  -H "Content-Type: application/json" \
  -d '{
    "answers": {
      "q1": "50万",
      "q2": "1年",
      "q3": "中等"
    }
  }'
```

### 辩论阶段

#### 开始辩论

```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/debate/start
```

#### 流式获取辩论内容 (SSE)

```bash
curl http://localhost:8080/api/v1/sessions/{session_id}/debate/stream
```

### 决议阶段

#### 生成决议书

```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/resolution/generate
```

#### 获取决议书

```bash
curl http://localhost:8080/api/v1/sessions/{session_id}/resolution
```

### 配置管理

#### 获取可用智者列表

```bash
curl http://localhost:8080/api/v1/config/agents
```

响应：
```json
{
  "agents": [
    {
      "id": "zhuge_liang",
      "name": "诸葛亮",
      "role": "首席战略官",
      "description": "蜀汉丞相，擅长战略规划与全局思考"
    },
    {
      "id": "wei_zheng",
      "name": "魏征",
      "role": "风险评估专家",
      "description": "唐代名臣，以直言进谏著称"
    }
  ]
}
```

## 环境变量说明

| 变量 | 说明 | 默认值 |
|------|------|--------|
| SERVER_PORT | 服务器端口 | 8080 |
| DB_PATH | SQLite 数据库路径 | ./data/sage_roundtable.db |
| CACHE_TTL | 缓存过期时间（秒） | 3600 |
| RATE_LIMIT | 速率限制（每分钟请求数） | 60 |
| LOG_LEVEL | 日志级别 | debug |
| LOG_FORMAT | 日志格式 | text |

## 架构说明

### 核心组件

1. **API 层** (`internal/api/`)
   - HTTP handlers
   - Middleware (CORS, Logger, Recovery, RateLimit)
   - Router (基于 chi)

2. **会话存储** (`internal/session/store.go`)
   - SQLite 持久化
   - 完整的 CRUD 操作
   - 关联数据加载

3. **缓存系统** (`pkg/cache/cache.go`)
   - 内存缓存
   - 自动过期清理
   - 线程安全

4. **错误处理** (`pkg/errors/errors.go`)
   - 统一错误码
   - HTTP 状态码映射
   - 结构化错误响应

### 数据库 Schema

- `sessions`: 会话主表
- `questions`: 问题列表
- `debates`: 辩论轮次
- `statements`: 智者发言
- `resolutions`: 决议书

所有表都通过外键关联，支持级联删除。

## 测试

运行单元测试：

```bash
go test ./internal/api/... -v
```

## 下一步

Phase 1 已完成以下功能：
- ✅ API 层实现
- ✅ SQLite 持久化
- ✅ 缓存系统
- ✅ 错误处理
- ✅ 速率限制

待实现功能：
- [ ] 集成 Questioning Engine
- [ ] 集成 Debate Engine
- [ ] 集成 Resolution Generator
- [ ] 实现 SSE 流式输出
- [ ] Web 前端界面
