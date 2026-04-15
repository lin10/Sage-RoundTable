package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode string

const (
	// 客户端错误
	ErrInvalidInput    ErrorCode = "INVALID_INPUT"
	ErrValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrNotFound        ErrorCode = "NOT_FOUND"
	ErrUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrForbidden       ErrorCode = "FORBIDDEN"
	
	// 业务逻辑错误
	ErrSessionNotFound ErrorCode = "SESSION_NOT_FOUND"
	ErrAgentNotFound   ErrorCode = "AGENT_NOT_FOUND"
	ErrSkillNotFound   ErrorCode = "SKILL_NOT_FOUND"
	ErrFlowNotFound    ErrorCode = "FLOW_NOT_FOUND"
	
	// LLM 相关错误
	ErrLLMAPIError     ErrorCode = "LLM_API_ERROR"
	ErrLLMTimeout      ErrorCode = "LLM_TIMEOUT"
	ErrTokenLimit      ErrorCode = "TOKEN_LIMIT_EXCEEDED"
	
	// 数据库错误
	ErrDatabaseError   ErrorCode = "DATABASE_ERROR"
	ErrQueryFailed     ErrorCode = "QUERY_FAILED"
	
	// 系统错误
	ErrInternalError   ErrorCode = "INTERNAL_ERROR"
	ErrRateLimit       ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// AppError 应用级错误结构
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAppError 创建应用错误
func NewAppError(code ErrorCode, message string, details string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: getDefaultHTTPStatus(code),
	}
}

// WithHTTPStatus 设置 HTTP 状态码
func (e *AppError) WithHTTPStatus(status int) *AppError {
	e.HTTPStatus = status
	return e
}

// getDefaultHTTPStatus 根据错误码返回默认 HTTP 状态码
func getDefaultHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrInvalidInput, ErrValidationFailed:
		return http.StatusBadRequest
	case ErrNotFound, ErrSessionNotFound, ErrAgentNotFound, ErrSkillNotFound, ErrFlowNotFound:
		return http.StatusNotFound
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrRateLimit:
		return http.StatusTooManyRequests
	case ErrLLMTimeout:
		return http.StatusGatewayTimeout
	case ErrServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrInternalError, ErrDatabaseError, ErrQueryFailed:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// IsAppError 判断是否为应用错误
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError 转换为应用错误
func AsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// WrapError 包装错误为应用错误
func WrapError(code ErrorCode, err error, message string) *AppError {
	if err == nil {
		return nil
	}
	
	details := err.Error()
	return NewAppError(code, message, details)
}
