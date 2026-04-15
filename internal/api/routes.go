package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router 路由器
type Router struct {
	chi.Router
}

// NewRouter 创建新的路由器
func NewRouter(handler *Handler) *Router {
	r := chi.NewRouter()
	
	// 全局中间件
	r.Use(middleware.Recoverer)
	r.Use(Logger)
	r.Use(CORS)
	
	// 注册 API 路由
	handler.RegisterRoutes(r)
	
	// 健康检查端点
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})
	
	return &Router{Router: r}
}
