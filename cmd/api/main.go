package main

import (
	"context"
	"errors"
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
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	if err := config.Init(); err != nil {
		panic(err)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	hub := ws.NewHub()

	amqpConn, err := amqp.Dial(config.AMQP.ToURI())
	if err != nil {
		panic(err)
	}
	defer func() { _ = amqpConn.Close() }()

	queuePublisher, err := rabbitmq.NewPublisher(amqpConn, "messages:new")
	if err != nil {
		panic(err)
	}

	rcl := redis.NewRedisClient()
	pubsubSubscriber := redis.NewPubSubSubscriber(rcl)

	dbPool, err := postgres.NewPostgresPool(context.Background(), config.Postgres.ToURI())
	if err != nil {
		panic(err)
	}
	defer dbPool.Close()

	dbQueries := queries.New(dbPool)
	messageRepository := postgres.NewMessageRepository(dbQueries)

	chatService := service.NewChatService(hub, queuePublisher, pubsubSubscriber, messageRepository)
	chatHandler := handlers.NewChatHandler(chatService)

	ginRouter := gin.Default()
	api.RegisterRoutes(ginRouter, chatHandler)

	addr := ":" + strconv.Itoa(config.API.API_PORT)
	srv := &http.Server{
		Addr:              addr,
		Handler:           ginRouter,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := chatService.BroadcastMessage(ctx); err != nil {
			log.Printf("[api] broadcast exited: %v", err)
		}
	}()

	go func() {
		log.Printf("[api] listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[api] listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Print("[api] shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[api] server forced to shutdown: %v", err)
	}

	cancel()
}
