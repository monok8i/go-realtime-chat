package main

import (
	"context"
	"go-realtime-chat/internal/api"
	"go-realtime-chat/internal/api/handlers"
	"go-realtime-chat/internal/api/ws"
	"go-realtime-chat/internal/config"
	"go-realtime-chat/internal/infra/rabbitmq"
	"go-realtime-chat/internal/infra/redis"
	"go-realtime-chat/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

// main is the application entry point. It initialises the Gin router, registers
// routes, and starts the HTTP server on port 8080.
func main() {
	err := config.Init()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := gin.Default()

	hub := ws.NewHub()

	amqpConn, err := amqp.Dial(config.AMQP.ToURI())
	if err != nil {
		panic(err)
	}
	defer amqpConn.Close()

	queuePublisher, err := rabbitmq.NewPublisher(amqpConn, "messages:new")
	if err != nil {
		panic(err)
	}

	rcl := redis.NewRedisClient()
	pubsubSubscriber := redis.NewRedisPubSubSubscriber(rcl)

	chatService := service.NewChatService(hub, queuePublisher, pubsubSubscriber)
	chatHandler := handlers.NewChatHandler(chatService)

	api.RegisterRoutes(r, chatHandler)

	go chatService.BroadcastMessage(ctx)

	_ = r.Run(":" + strconv.Itoa(config.API.API_PORT))
}
