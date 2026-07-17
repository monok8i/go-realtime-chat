package main

import (
	"context"
	"go-realtime-chat/internal/config"
	"go-realtime-chat/internal/infra/rabbitmq"
	"go-realtime-chat/internal/infra/redis"
	"go-realtime-chat/internal/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

// main is the worker entry point. It consumes messages from queue and send it to storage and redis channel
func main() {
	err := config.Init()
	if err != nil {
		panic(err)
	}

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

	workerService := service.NewWorkerService(consumer, pubsubpublisher)

	go workerService.Consuming(ctx)
}
