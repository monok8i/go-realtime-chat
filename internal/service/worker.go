// Package service provides the core business logic for the chat system.
package service

import (
	"context"
	"encoding/json"
	"go-realtime-chat/internal/config"
	"go-realtime-chat/internal/domain"
	"log"
)

type messageRepository interface {
	CreateNewMessage(ctx context.Context, payload domain.Payload) error
}

// WorkerService consumes messages from a queue and republishes them to Redis PubSub.
type WorkerService struct {
	consumer        domain.QueueConsumer
	pubsubpublisher domain.PubSubPublisher
	repo            messageRepository
}

// NewWorkerService creates a new WorkerService.
func NewWorkerService(consumer domain.QueueConsumer, pubsubpublisher domain.PubSubPublisher, repo messageRepository) *WorkerService {
	return &WorkerService{
		consumer:        consumer,
		pubsubpublisher: pubsubpublisher,
		repo:            repo,
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

			var payload domain.Payload
			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				log.Printf("[worker] consume: unmarshal error: %v", err)
				if err := msg.Ack(); err != nil {
					log.Printf("[worker] consume: ack error: %v", err)
				}
				continue
			}

			if err := s.repo.CreateNewMessage(ctx, payload); err != nil {
				log.Printf("[worker] consume: save message error: %v", err)
				continue
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
