package main

import (
	"context"
	"go-realtime-chat/internal/config"
	"go-realtime-chat/internal/infra/postgres"
	"go-realtime-chat/internal/infra/postgres/gen"
	"go-realtime-chat/internal/infra/rabbitmq"
	"go-realtime-chat/internal/infra/redis"
	"go-realtime-chat/internal/service"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	if err := config.Init(); err != nil {
		panic(err)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	amqpConn, err := amqp.Dial(config.AMQP.ToURI())
	if err != nil {
		panic(err)
	}
	defer func() { _ = amqpConn.Close() }()

	consumer, err := rabbitmq.NewConsumer(amqpConn, "messages:new")
	if err != nil {
		panic(err)
	}

	rcl := redis.NewRedisClient()
	pubsubPublisher := redis.NewPubSubPublisher(rcl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := postgres.NewPostgresPool(ctx, config.Postgres.ToURI())
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	q := gen.New(pool)
	repo := postgres.NewMessageRepository(q)

	workerService := service.NewWorkerService(consumer, pubsubPublisher, repo)

	go func() {
		if err := workerService.Consuming(ctx); err != nil {
			log.Printf("[worker] consuming exited: %v", err)
		}
	}()

	log.Print("[worker] started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Print("[worker] shutting down...")
	cancel()
}
