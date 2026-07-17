package service

import (
	"context"
	"go-realtime-chat/internal/config"
	"go-realtime-chat/internal/domain"
	"log"
)

type WorkerService struct {
	consumer        domain.QueueConsumer
	pubsubpublisher domain.PubSubPublisher
	// repo            domain.MessageRepository
}

func NewWorkerService(consumer domain.QueueConsumer, pubsubpublisher domain.PubSubPublisher) *WorkerService {
	return &WorkerService{
		consumer:        consumer,
		pubsubpublisher: pubsubpublisher,
		// repo:            repo,
	}
}

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

			// var payload domain.Payload

			// if err := json.Unmarshal(msg.Body, &payload); err != nil {
			// 	log.Printf("Consuming: unmarshal error: %v", err)
			// 	continue
			// }

			// if err := s.repo.SaveMessage(ctx, domain.Payload{
			// 	ChatID:  payload.ChatID,
			// 	UserID:  payload.UserID,
			// 	Message: payload.Message,
			// }); err != nil {
			// 	log.Printf("Consuming: save error: %v", err)
			// 	continue
			// }

			if err := s.pubsubpublisher.Publish(ctx, config.Redis.PUBSUB_CHANNEL, msg.Body); err != nil {
				log.Printf("Consuming: pubsubpublisher error: %v", err)
				continue
			}
			if err := msg.Ack(); err != nil {
				log.Printf("Consuming: ack error: %v", err)
				continue
			}
		}
	}
}
