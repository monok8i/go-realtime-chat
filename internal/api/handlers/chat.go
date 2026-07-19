package handlers

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"go-realtime-chat/internal/api/ws"
	"go-realtime-chat/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var chatIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// ChatHandlerImpl handles HTTP and WebSocket requests for the chat application.
type ChatHandlerImpl struct {
	svc *service.ChatService
}

// NewChatHandler creates a new ChatHandlerImpl with the given chat service.
func NewChatHandler(svc *service.ChatService) *ChatHandlerImpl {
	return &ChatHandlerImpl{svc: svc}
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
		client.ReadPump(ctx, h.svc.HandleIncomingMessage)
		cancel()
		h.svc.RemoveClient(client)
	}()
}

const maxLimit = 1000

// GetMessagesByChat returns paginated messages for the specified chat ID.
// Query params: limit (default 50, max 1000), offset (default 0).
func (h *ChatHandlerImpl) GetMessagesByChat(c *gin.Context) {
	chatID := c.Param("chat_id")
	if !chatIDPattern.MatchString(chatID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 {
		limit = 50
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	resp, err := h.svc.GetMessagesByChat(c.Request.Context(), chatID, limit, offset)
	if err != nil {
		log.Printf("[handlers] get messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
