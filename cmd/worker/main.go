// Package main is the entry point for the Worker binary.
//
// It consumes messages from RabbitMQ and republishes them to Redis PubSub
// for cross-instance message broadcasting.
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

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	err := config.Init()
	if err != nil {
		panic(err)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	amqpConn, err := amqp.Dial(config.AMQP.ToURI())
	if err != nil {
		panic(err)
	}

	consumer, err := rabbitmq.NewRabbitMQConsumer(amqpConn, "messages:new")
	if err != nil {
		panic(err)
	}

	rcl := redis.NewRedisClient()
	pubsubpublisher := redis.NewRedisPubSubPublisher(rcl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := postgres.NewPostgresPool(ctx, config.Postgres.ToURI())
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	q := gen.New(pool)
	repo := postgres.NewMessageRepository(q)

	workerService := service.NewWorkerService(consumer, pubsubpublisher, repo)

	go workerService.Consuming(ctx)
}
