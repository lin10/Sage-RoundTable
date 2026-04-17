package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sage-roundtable/core/internal/api"
	"github.com/sage-roundtable/core/internal/session"
	"github.com/sage-roundtable/core/pkg/cache"
	"github.com/sage-roundtable/core/pkg/logger"
)

func main() {
	// 1. 初始化日志
	logger.Init("debug", "text")
	slog.Info("Starting Sage-RoundTable Server...")

	// 2. 加载配置 (从环境变量或配置文件)
	serverPort := getEnv("SERVER_PORT", "8080")
	cacheTTL := getEnvAsInt("CACHE_TTL", 3600) // 默认 1 小时
	rateLimit := getEnvAsInt("RATE_LIMIT", 60) // 默认每分钟 60 次请求

	// 3. 初始化内存存储（MVP 版本，无需数据库）
	store := session.NewMemoryStore()
	slog.Info("Memory store initialized")

	// 4. 初始化缓存
	appCache := cache.NewCache(time.Duration(cacheTTL) * time.Second)
	slog.Info("Cache initialized", "ttl_seconds", cacheTTL)

	// 5. 创建 API 处理器
	handler := api.NewHandler(store, appCache)

	// 6. 创建路由器并添加中间件（必须在注册路由之前）
	router := chi.NewRouter()

	// 添加全局中间件
	rateLimiter := api.NewRateLimiter(rateLimit, time.Minute)
	router.Use(rateLimiter.Middleware)

	// 注册 API 路由
	handler.RegisterRoutes(router)

	// 7. 启动 HTTP 服务器
	addr := fmt.Sprintf(":%s", serverPort)
	slog.Info("Server starting", "address", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		var value int
		if _, err := fmt.Sscanf(valueStr, "%d", &value); err == nil {
			return value
		}
	}
	return defaultValue
}
