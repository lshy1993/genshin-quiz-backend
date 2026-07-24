package common

import (
	"fmt"
	"net/http"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error %d: %s", e.Code, e.Message)
}

func NewBadRequestError(message string) *APIError {
	// HTTP 状态码对应的错误构造函数
	return &APIError{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

func NewUnauthorizedError(message string) *APIError {
	return &APIError{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

func NewNotFoundError(message string) *APIError {
	return &APIError{
		Code:    http.StatusNotFound,
		Message: message,
	}
}

func NewInternalServerError(message string) *APIError {
	return &APIError{
		Code:    http.StatusInternalServerError,
		Message: message,
	}
}

var (
	ErrNotFound           = NewNotFoundError("记录不存在")
	ErrUserNotFound       = NewNotFoundError("用户不存在")
	ErrUserAlreadyExists  = NewBadRequestError("用户已存在")
	ErrInvalidCredentials = NewUnauthorizedError("邮箱或密码错误")
	ErrQuestionNotFound   = NewNotFoundError("问题未找到")
	ErrVoteNotFound       = NewNotFoundError("投票未找到")
	ErrInvalidToken       = NewUnauthorizedError("Invalid or expired token")
	ErrInvalidTokenFormat = NewUnauthorizedError("Invalid authorization header format")
	ErrUserNotInContext   = NewUnauthorizedError("用户未登录或认证失败")
	ErrDatabaseError      = NewInternalServerError("Database error")
	ErrAdminAuthError     = NewInternalServerError("Admin access required")
)
