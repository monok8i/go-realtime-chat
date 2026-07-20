// Package api provides HTTP route registration for the chat service.
package api

import (
	"context"
	"net/http"

	"go-realtime-chat/internal/api/handlers"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
)

// ChatHandler defines the contract for HTTP and WebSocket handlers.
type ChatHandler interface {
	Health(ctx context.Context, input *struct{}) (*handlers.HealthOutput, error)
	GetMessagesByChat(ctx context.Context, input *handlers.GetMessagesInput) (*handlers.GetMessagesOutput, error)
	HandleWebSocket(w http.ResponseWriter, r *http.Request)
}

// RegisterRoutes registers all application HTTP routes on the provided Huma API and chi router.
func RegisterRoutes(api huma.API, r chi.Router, h ChatHandler) {
	r.Get("/api/ws/chat", h.HandleWebSocket)

	huma.Register(api, huma.Operation{
		OperationID: "health",
		Summary:     "Health check",
		Method:      http.MethodGet,
		Path:        "/api/health",
	}, h.Health)

	huma.Register(api, huma.Operation{
		OperationID: "getMessages",
		Summary:     "Get paginated chat messages",
		Method:      http.MethodGet,
		Path:        "/api/chats/{chat_id}/messages",
	}, h.GetMessagesByChat)
}
