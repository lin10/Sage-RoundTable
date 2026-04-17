package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sage-roundtable/core/internal/session"
	"github.com/stretchr/testify/assert"
)

// setupTestRouter 设置测试路由器
func setupTestRouter() (*Router, *session.SQLiteStore) {
	// 使用内存数据库进行测试
	store, err := session.NewSQLiteStore(":memory:")
	if err != nil {
		panic(err)
	}

	handler := NewHandler(store)
	router := NewRouter(handler)

	return router, store
}

func TestHealthEndpoint(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "healthy", response["status"])
}

func TestCreateSession(t *testing.T) {
	router, store := setupTestRouter()
	defer store.Close()

	payload := CreateSessionRequest{
		InitialQuery: "我应该辞职创业吗？",
		UserID:       "test_user",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response CreateSessionResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response.SessionID)
	assert.Equal(t, session.SessionStateInitialized, response.State)
}

func TestGetSession(t *testing.T) {
	router, store := setupTestRouter()
	defer store.Close()

	// 先创建会话
	sess := &session.Session{
		ID:           "test_session_1",
		UserID:       "test_user",
		State:        session.SessionStateInitialized,
		InitialQuery: "测试问题",
	}
	store.CreateSession(sess)

	// 获取会话
	req := httptest.NewRequest("GET", "/api/v1/sessions/test_session_1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response session.Session
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test_session_1", response.ID)
	assert.Equal(t, "测试问题", response.InitialQuery)
}

func TestListSessions(t *testing.T) {
	router, store := setupTestRouter()
	defer store.Close()

	// 创建多个会话
	for i := 0; i < 3; i++ {
		sess := &session.Session{
			ID:           "test_session_" + string(rune('0'+i)),
			UserID:       "test_user",
			State:        session.SessionStateInitialized,
			InitialQuery: "测试问题",
		}
		store.CreateSession(sess)
	}

	// 列出会话
	req := httptest.NewRequest("GET", "/api/v1/sessions?user_id=test_user", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["sessions"])
}

func TestDeleteSession(t *testing.T) {
	router, store := setupTestRouter()
	defer store.Close()

	// 先创建会话
	sess := &session.Session{
		ID:           "test_session_delete",
		UserID:       "test_user",
		State:        session.SessionStateInitialized,
		InitialQuery: "测试问题",
	}
	store.CreateSession(sess)

	// 删除会话
	req := httptest.NewRequest("DELETE", "/api/v1/sessions/test_session_delete", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// 验证已删除
	_, err := store.GetSession("test_session_delete")
	assert.Error(t, err)
}
