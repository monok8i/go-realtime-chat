// Package main is the entry point for the API server binary.
//
// It initialises the Gin HTTP server with WebSocket support,
// connects to RabbitMQ and Redis, and starts the broadcast listener.
package main

import (
	"context"
	"go-realtime-chat/internal/api"
	"go-realtime-chat/internal/api/handlers"
	"go-realtime-chat/internal/api/ws"
	"go-realtime-chat/internal/config"
	"go-realtime-chat/internal/infra/postgres"
	queries "go-realtime-chat/internal/infra/postgres/gen"
	"go-realtime-chat/internal/infra/rabbitmq"
	"go-realtime-chat/internal/infra/redis"
	"go-realtime-chat/internal/service"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	err := config.Init()
	if err != nil {
		panic(err)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

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

	dbPool, err := postgres.NewPostgresPool(ctx, config.Postgres.ToURI())
	if err != nil {
		panic(err)
	}
	dbQueries := queries.New(dbPool)
	messageRepository := postgres.NewMessageRepository(dbQueries)

	chatService := service.NewChatService(hub, queuePublisher, pubsubSubscriber, messageRepository)
	chatHandler := handlers.NewChatHandler(chatService)

	api.RegisterRoutes(r, chatHandler)

	go chatService.BroadcastMessage(ctx)

	_ = r.Run(":" + strconv.Itoa(config.API.API_PORT))
}
