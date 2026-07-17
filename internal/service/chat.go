package service

import (
	"context"
	"encoding/json"
	"go-realtime-chat/internal/config"
	"go-realtime-chat/internal/domain"
	"log"
)

type ChatService struct {
	hub              domain.Hub
	publisher        domain.QueuePublisher
	pubsubsubscriber domain.PubSubSubscriber
}

func NewChatService(hub domain.Hub, publisher domain.QueuePublisher, pubsubsubscriber domain.PubSubSubscriber) *ChatService {
	return &ChatService{hub: hub, publisher: publisher, pubsubsubscriber: pubsubsubscriber}
}

func (cs *ChatService) AddClient(c domain.Client, chatId string) {
	cs.hub.AddClient(c, chatId)
}

func (cs *ChatService) RemoveClient(c domain.Client) {
	cs.hub.RemoveClient(c)
}

func (cs *ChatService) PublishToBroker(ctx context.Context, payload domain.Payload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("PublishToBroker: marshal error: %v", err)
		return err
	}

	err = cs.publisher.Publish(ctx, body)
	if err != nil {
		log.Printf("PublishToBroker: amqp error: %v", err)
		return err
	}

	return nil
}

func (cs *ChatService) HandleIncomingMessage(ctx context.Context, c domain.Client, payload domain.Payload) error {
	cs.AddClient(c, payload.ChatID)

	if err := cs.PublishToBroker(ctx, payload); err != nil {
		log.Printf("HandleIncomingMessage: failed to send message to broker: %v", err)
		return err
	}

	return nil
}

func (cs *ChatService) BroadcastMessage(ctx context.Context) error {
	out, err := cs.pubsubsubscriber.Subscribe(ctx, config.Redis.PUBSUB_CHANNEL)
	if err != nil {
		log.Printf("Failed to subscribe to redis channel: %v", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-out:
			bytep := msg.Payload

			var payload domain.Payload
			if err := json.Unmarshal(bytep, &payload); err != nil {
				log.Printf("readPump: unmarshal error: %v", err)
				continue
			}

			cs.hub.Broadcast(payload.ChatID, payload)
		}
	}

}
