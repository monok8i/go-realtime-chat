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

type ChatHandlerImpl struct {
	service domain.ChatService
}

func NewChatHandler(service domain.ChatService) *ChatHandlerImpl {
	return &ChatHandlerImpl{service: service}
}

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *ChatHandlerImpl) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket error: %v", err)
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
