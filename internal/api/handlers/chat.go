package handlers

import (
	"context"
	"log"
	"net/http"

	"go-realtime-chat/internal/api/ws"
	"go-realtime-chat/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ChatHandlerImpl handles HTTP and WebSocket requests for the chat application.
type ChatHandlerImpl struct {
	service domain.ChatService
}

// NewChatHandler creates a new ChatHandlerImpl with the given chat service.
func NewChatHandler(service domain.ChatService) *ChatHandlerImpl {
	return &ChatHandlerImpl{service: service}
}

// Health responds with a simple health check status.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket upgrades the HTTP connection to WebSocket and manages the client lifecycle.
//
// It creates a background context for the connection, starts the write and read pumps,
// and ensures the client is removed from the hub when the connection closes.
func (h *ChatHandlerImpl) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[handlers] websocket upgrade error: %v", err)
		return
	}
	client := ws.NewClient(conn)

	ctx, cancel := context.WithCancel(context.Background())

	go client.WritePump(ctx)

	go func() {
		client.ReadPump(ctx, h.service.HandleIncomingMessage)
		h.service.RemoveClient(client)
		cancel()
	}()
}
