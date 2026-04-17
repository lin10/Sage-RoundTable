package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sage-roundtable/core/internal/session"
	"github.com/sage-roundtable/core/pkg/cache"
	apperrors "github.com/sage-roundtable/core/pkg/errors"
	"github.com/sage-roundtable/core/pkg/logger"
)

// Handler API 处理器
type Handler struct {
	store session.Store
	cache *cache.Cache
}

// NewHandler 创建处理器实例
func NewHandler(store session.Store, cache *cache.Cache) *Handler {
	return &Handler{store: store, cache: cache}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1", func(r chi.Router) {
		// 会话管理
		r.Post("/sessions", h.CreateSession)
		r.Get("/sessions", h.ListSessions)
		r.Get("/sessions/{id}", h.GetSession)
		r.Delete("/sessions/{id}", h.DeleteSession)

		// 提问阶段
		r.Post("/sessions/{id}/questions", h.GenerateQuestions)
		r.Post("/sessions/{id}/answers", h.SubmitAnswers)

		// 辩论阶段
		r.Post("/sessions/{id}/debate/start", h.StartDebate)
		r.Get("/sessions/{id}/debate/stream", h.StreamDebate)

		// 决议阶段
		r.Post("/sessions/{id}/resolution/generate", h.GenerateResolution)
		r.Get("/sessions/{id}/resolution", h.GetResolution)

		// 配置管理
		r.Get("/config/agents", h.ListAgents)
	})
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	InitialQuery string            `json:"initial_query"`
	UserID       string            `json:"user_id"`
	AgentIDs     []string          `json:"agent_ids,omitempty"`
	ContextInfo  map[string]string `json:"context_info,omitempty"`
}

// CreateSessionResponse 创建会话响应
type CreateSessionResponse struct {
	SessionID string               `json:"session_id"`
	State     session.SessionState `json:"state"`
}

// CreateSession 创建新会话
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, apperrors.NewAppError(apperrors.ErrInvalidInput, "Invalid request body", err.Error()))
		return
	}

	if req.InitialQuery == "" {
		WriteError(w, apperrors.NewAppError(apperrors.ErrValidationFailed, "initial_query is required", ""))
		return
	}

	if req.UserID == "" {
		req.UserID = "anonymous"
	}

	// TODO: 生成会话 ID，初始化会话对象并存储
	sessionID := "sess_" + req.InitialQuery[:8] // 临时实现

	sess := &session.Session{
		ID:           sessionID,
		UserID:       req.UserID,
		State:        session.SessionStateInitialized,
		InitialQuery: req.InitialQuery,
		ContextInfo:  req.ContextInfo,
	}

	if err := h.store.CreateSession(sess); err != nil {
		WriteError(w, err)
		return
	}

	logger.Log.Info("Session created", "id", sessionID, "query", req.InitialQuery)

	WriteJSON(w, http.StatusCreated, CreateSessionResponse{
		SessionID: sessionID,
		State:     sess.State,
	})
}

// GetSession 获取会话详情
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	sess, err := h.store.GetSession(id)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, sess)
}

// ListSessions 列出会话
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	limit := 20
	offset := 0

	sessions, err := h.store.ListSessions(userID, limit, offset)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// DeleteSession 删除会话
func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.store.DeleteSession(id); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GenerateQuestions 生成追问问题
func (h *Handler) GenerateQuestions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// TODO: 调用 questioning engine 生成问题
	sess, err := h.store.GetSession(id)
	if err != nil {
		WriteError(w, err)
		return
	}

	// 临时返回示例问题
	questions := []string{
		"您的预算范围是多少？",
		"您期望的时间周期是多久？",
		"您的风险承受能力如何？",
	}

	// 更新会话状态
	sess.State = session.SessionStateQuestioning
	h.store.UpdateSession(sess)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"questions": questions,
	})
}

// SubmitAnswersRequest 提交答案请求
type SubmitAnswersRequest struct {
	Answers map[string]string `json:"answers"`
}

// SubmitAnswers 提交问题答案
func (h *Handler) SubmitAnswers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req SubmitAnswersRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, apperrors.NewAppError(apperrors.ErrInvalidInput, "Invalid request body", err.Error()))
		return
	}

	// TODO: 保存答案到数据库
	// 更新会话状态
	sess, err := h.store.GetSession(id)
	if err != nil {
		WriteError(w, err)
		return
	}

	sess.State = session.SessionStateDebating
	h.store.UpdateSession(sess)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"state": sess.State,
	})
}

// StartDebate 开始辩论
func (h *Handler) StartDebate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// TODO: 调用 debate engine
	sess, err := h.store.GetSession(id)
	if err != nil {
		WriteError(w, err)
		return
	}

	sess.State = session.SessionStateDebating
	h.store.UpdateSession(sess)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"state":   sess.State,
		"message": "Debate started",
	})
}

// StreamDebate 流式获取辩论内容 (SSE)
func (h *Handler) StreamDebate(w http.ResponseWriter, r *http.Request) {
	// TODO: 实现 Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// 发送示例事件
	data := `{"agent_id": "zhuge_liang", "agent_name": "诸葛亮", "content": "臣以为...", "timestamp": "2026-04-15T10:30:00Z"}`
	fmt.Fprintf(w, "event: statement\ndata: %s\n\n", data)
	flusher.Flush()
}

// GenerateResolution 生成决议书
func (h *Handler) GenerateResolution(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// TODO: 调用 resolution generator
	sess, err := h.store.GetSession(id)
	if err != nil {
		WriteError(w, err)
		return
	}

	sess.State = session.SessionStateResolving
	h.store.UpdateSession(sess)

	// 临时返回示例决议
	resolution := map[string]interface{}{
		"consensus": []string{"共识点1", "共识点2"},
		"strategies": map[string]interface{}{
			"upper":  "上策方案",
			"middle": "中策方案",
			"lower":  "下策方案",
		},
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"resolution": resolution,
		"state":      sess.State,
	})
}

// GetResolution 获取决议书
func (h *Handler) GetResolution(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	resolution, err := h.store.GetResolution(id)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, resolution)
}

// ListAgents 获取可用智者列表
func (h *Handler) ListAgents(w http.ResponseWriter, r *http.Request) {
	// TODO: 从 agent registry 加载
	agents := []map[string]interface{}{
		{
			"id":          "zhuge_liang",
			"name":        "诸葛亮",
			"role":        "首席战略官",
			"description": "蜀汉丞相，擅长战略规划与全局思考",
		},
		{
			"id":          "wei_zheng",
			"name":        "魏征",
			"role":        "风险评估专家",
			"description": "唐代名臣，以直言进谏著称",
		},
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"agents": agents,
	})
}
