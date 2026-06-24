// Package httperrors cung cấp kiểu lỗi mang sẵn HTTP status code và helper
// để handler ánh xạ error -> status code mà không cần phụ thuộc gRPC codes.
package httperrors

import (
	"errors"
	"net/http"
)

// ResponseError là body JSON trả về cho client khi có lỗi.
// Lưu ý: json tag là "error" để khớp với frontend hiện tại (đọc data.error).
type ResponseError struct {
	Message string `json:"error"`
}

// HTTPError là lỗi nghiệp vụ kèm HTTP status code.
type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

func newError(code int, msg string) *HTTPError {
	return &HTTPError{Code: code, Message: msg}
}

func NewBadRequest(msg string) *HTTPError   { return newError(http.StatusBadRequest, msg) }
func NewUnauthorized(msg string) *HTTPError { return newError(http.StatusUnauthorized, msg) }
func NewForbidden(msg string) *HTTPError    { return newError(http.StatusForbidden, msg) }
func NewNotFound(msg string) *HTTPError     { return newError(http.StatusNotFound, msg) }
func NewConflict(msg string) *HTTPError     { return newError(http.StatusConflict, msg) }
func NewInternal(msg string) *HTTPError     { return newError(http.StatusInternalServerError, msg) }

// GetStatusCode trả HTTP status code tương ứng với error.
// Nếu không phải *HTTPError thì coi là lỗi nội bộ (500).
func GetStatusCode(err error) int {
	var he *HTTPError
	if errors.As(err, &he) {
		return he.Code
	}
	return http.StatusInternalServerError
}
