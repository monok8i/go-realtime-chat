// Package service provides the core business logic for the chat system.
package service

import (
	"context"
	"go-realtime-chat/internal/config"
	"go-realtime-chat/internal/domain"
	"log"
)

// WorkerService consumes messages from a queue and republishes them to Redis PubSub.
type WorkerService struct {
	consumer        domain.QueueConsumer
	pubsubpublisher domain.PubSubPublisher
}

// NewWorkerService creates a new WorkerService.
func NewWorkerService(consumer domain.QueueConsumer, pubsubpublisher domain.PubSubPublisher) *WorkerService {
	return &WorkerService{
		consumer:        consumer,
		pubsubpublisher: pubsubpublisher,
	}
}

// Consuming reads messages from the queue and publishes them to Redis PubSub.
// It acknowledges each message after successful processing.
func (s *WorkerService) Consuming(ctx context.Context) error {
	messages, err := s.consumer.Consume(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg, ok := <-messages:
			if !ok {
				return nil
			}

			if err := s.pubsubpublisher.Publish(ctx, config.Redis.PUBSUB_CHANNEL, msg.Body); err != nil {
				log.Printf("[worker] consume: publish to redis error: %v", err)
				continue
			}
			if err := msg.Ack(); err != nil {
				log.Printf("[worker] consume: ack error: %v", err)
				continue
			}
		}
	}
}
