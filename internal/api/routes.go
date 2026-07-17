package api

import (
	"go-realtime-chat/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

// ChatHandler defines the contract for chat-related HTTP handlers.
type ChatHandler interface {
	HandleWebSocket(c *gin.Context)
}

// RegisterRoutes registers all application HTTP routes on the provided Gin engine.
func RegisterRoutes(r *gin.Engine, ch ChatHandler) {

	{
		router := r.Group("/api")
		router.GET("/health", handlers.Health)
		router.GET("/ws/chat", ch.HandleWebSocket)
	}
}
