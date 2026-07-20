// Package handlers provides HTTP and WebSocket request handlers for the chat API.
package handlers

import (
	"context"
	"log"
	"net/http"

	"go-realtime-chat/internal/api/ws"
	"go-realtime-chat/internal/domain"
	"go-realtime-chat/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gorilla/websocket"
)

// ChatHandlerImpl handles HTTP and WebSocket requests for the chat application.
type ChatHandlerImpl struct {
	svc *service.ChatService
}

// NewChatHandler creates a new ChatHandlerImpl with the given chat service.
func NewChatHandler(svc *service.ChatService) *ChatHandlerImpl {
	return &ChatHandlerImpl{svc: svc}
}

// HealthOutput is the response body for the health endpoint.
type HealthOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

// Health responds with a simple health check status.
func (h *ChatHandlerImpl) Health(ctx context.Context, input *struct{}) (*HealthOutput, error) {
	resp := &HealthOutput{}
	resp.Body.Status = "ok"
	return resp, nil
}

// GetMessagesInput is the request parameters for retrieving chat messages.
type GetMessagesInput struct {
	ChatID string `path:"chat_id" maxLength:"64"`
	Limit  int    `query:"limit" minimum:"1" maximum:"1000"`
	Offset int    `query:"offset" minimum:"0"`
}

// GetMessagesOutput is the paginated response for the messages endpoint.
type GetMessagesOutput struct {
	Body domain.GetMessagesResponse
}

// GetMessagesByChat returns paginated messages for the specified chat ID.
func (h *ChatHandlerImpl) GetMessagesByChat(ctx context.Context, input *GetMessagesInput) (*GetMessagesOutput, error) {
	limit := input.Limit
	if limit < 1 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}
	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	resp, err := h.svc.GetMessagesByChat(ctx, input.ChatID, limit, offset)
	if err != nil {
		log.Printf("[handlers] get messages: %v", err)
		return nil, huma.Error500InternalServerError("failed to get messages")
	}

	return &GetMessagesOutput{Body: *resp}, nil
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket upgrades the HTTP connection to WebSocket and manages the client lifecycle.
func (h *ChatHandlerImpl) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[handlers] websocket upgrade error: %v", err)
		return
	}
	client := ws.NewClient(conn)

	ctx, cancel := context.WithCancel(context.Background())

	go client.WritePump(ctx)

	go func() {
		client.ReadPump(ctx, h.svc.HandleIncomingMessage)
		cancel()
		h.svc.RemoveClient(client)
	}()
}
