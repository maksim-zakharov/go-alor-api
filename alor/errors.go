package alor

import (
	"errors"
	"fmt"
	"net/http"
)

// Типовые ошибки клиента.
var (
	ErrNoToken      = errors.New("alor: нет refresh/access токена")
	ErrUnauthorized = errors.New("alor: не авторизован")
	ErrRateLimited  = errors.New("alor: превышен лимит запросов")
	ErrClosed       = errors.New("alor: клиент закрыт")
	ErrNotConnected = errors.New("alor: WebSocket не подключён")
)

// APIError — ответ HTTP API с ненулевым статусом.
type APIError struct {
	/** StatusCode HTTP-код ответа. */
	StatusCode int
	/** Method HTTP-метод. */
	Method string
	/** Path путь запроса. */
	Path string
	/** Body тело ответа (обрезанное). */
	Body string
	/** Message текстовое сообщение из JSON, если удалось разобрать. */
	Message string
}

// Error реализует error.
func (e *APIError) Error() string {
	if e == nil {
		return "alor: api error"
	}
	msg := e.Message
	if msg == "" {
		msg = e.Body
	}
	if len(msg) > 300 {
		msg = msg[:300] + "…"
	}
	return fmt.Sprintf("alor: %s %s → %d: %s", e.Method, e.Path, e.StatusCode, msg)
}

// Temporary сообщает, имеет ли смысл повторить запрос.
func (e *APIError) Temporary() bool {
	if e == nil {
		return false
	}
	return e.StatusCode == http.StatusTooManyRequests ||
		e.StatusCode == http.StatusBadGateway ||
		e.StatusCode == http.StatusServiceUnavailable ||
		e.StatusCode == http.StatusGatewayTimeout
}

// WSError — ошибка подписки/команды WebSocket.
type WSError struct {
	/** GUID идентификатор запроса/подписки. */
	GUID string
	/** Code код операции или HTTP-подобный код из ответа. */
	Code int
	/** Message текст ошибки. */
	Message string
}

// Error реализует error.
func (e *WSError) Error() string {
	if e == nil {
		return "alor: websocket error"
	}
	if e.GUID != "" {
		return fmt.Sprintf("alor: ws guid=%s code=%d: %s", e.GUID, e.Code, e.Message)
	}
	return fmt.Sprintf("alor: ws code=%d: %s", e.Code, e.Message)
}
