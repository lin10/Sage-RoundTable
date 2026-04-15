package api

import (
	"net/http"
	"sync"
	"time"

	apperrors "github.com/sage-roundtable/core/pkg/errors"
)

// RateLimiter 简单的基于 IP 的速率限制器
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	
	// 启动定期清理任务
	go rl.cleanupLoop()
	
	return rl
}

// Middleware 速率限制中间件
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取客户端 IP
		ip := r.RemoteAddr
		
		if !rl.Allow(ip) {
			WriteError(w, apperrors.NewAppError(
				apperrors.ErrRateLimit,
				"Rate limit exceeded",
				"Too many requests. Please try again later.",
			).WithHTTPStatus(http.StatusTooManyRequests))
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	windowStart := now.Add(-rl.window)
	
	// 获取该 IP 的请求历史
	timestamps, exists := rl.requests[ip]
	if !exists {
		timestamps = []time.Time{}
	}
	
	// 过滤掉窗口外的请求
	validTimestamps := []time.Time{}
	for _, ts := range timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	
	// 检查是否超过限制
	if len(validTimestamps) >= rl.limit {
		rl.requests[ip] = validTimestamps
		return false
	}
	
	// 添加新请求
	validTimestamps = append(validTimestamps, now)
	rl.requests[ip] = validTimestamps
	
	return true
}

// cleanupLoop 定期清理过期的请求记录
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup 清理所有过期的请求记录
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	for ip, timestamps := range rl.requests {
		validTimestamps := []time.Time{}
		for _, ts := range timestamps {
			if ts.After(now.Add(-rl.window)) {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		
		if len(validTimestamps) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = validTimestamps
		}
	}
}
