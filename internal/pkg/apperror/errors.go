// Package apperror содержит типы и конструкторы ошибок приложения,
// обеспечивающие единый формат машиночитаемых ошибок для API и логирования.
package apperror

import "fmt"

// AppError представляет ошибку приложения с машиночитаемым кодом,
// человекочитаемым сообщением и HTTP-статусом для ответа.
type AppError struct {
	Code       string // машиночитаемый код ошибки, например "not_found"
	Message    string // человекочитаемое сообщение
	HTTPStatus int    // HTTP-статус ответа, например 404
}

// Error реализует интерфейс error.
func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError создаёт новую ошибку приложения с указанными кодом, сообщением и HTTP-статусом.
func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

// Предопределённые ошибки, используемые во всех слоях приложения.
var (
	// ErrNotFound используется, когда ресурс не найден.
	ErrNotFound = NewAppError("not_found", "resource not found", 404)
	// ErrInvalidInput используется при невалидных входных данных запроса.
	ErrInvalidInput = NewAppError("invalid_input", "invalid input", 400)
	// ErrConflict используется при конфликте состояния (например, попытка привязать неготовое медиа).
	ErrConflict = NewAppError("conflict", "conflict", 409)
	// ErrInternal используется для неожиданных внутренних ошибок.
	ErrInternal = NewAppError("internal_error", "internal server error", 500)
)
