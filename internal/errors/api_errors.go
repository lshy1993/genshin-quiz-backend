package errors

import (
	"fmt"
	"net/http"
)

// APIError 定义统一的 API 错误结构
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error %d: %s", e.Code, e.Message)
}

// HTTP 状态码对应的错误构造函数
func NewBadRequestError(message string) *APIError {
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

// 预定义的常用错误
var (
	ErrUserNotFound       = NewNotFoundError("用户不存在")
	ErrUserAlreadyExists  = NewBadRequestError("用户已存在")
	ErrInvalidCredentials = NewUnauthorizedError("邮箱或密码错误")
)
