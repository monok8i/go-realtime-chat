// Package service provides the core business logic for the chat system.
package service

import (
	"context"
	"encoding/json"
	"go-realtime-chat/internal/config"
	"go-realtime-chat/internal/domain"
	"go-realtime-chat/internal/infra/postgres"
	"log"
)

// ChatService implements the domain.ChatService interface.
// It orchestrates the flow of messages between WebSocket clients, RabbitMQ, and Redis PubSub.
type ChatService struct {
	hub              domain.Hub
	publisher        domain.QueuePublisher
	pubsubsubscriber domain.PubSubSubscriber
	repo             *postgres.MessageRepository
}

// NewChatService creates a new ChatService.
func NewChatService(hub domain.Hub, publisher domain.QueuePublisher, pubsubsubscriber domain.PubSubSubscriber, repo *postgres.MessageRepository) *ChatService {
	return &ChatService{hub: hub, publisher: publisher, pubsubsubscriber: pubsubsubscriber, repo: repo}
}

// AddClient registers a client in the given chat room.
func (cs *ChatService) AddClient(c domain.Client, chatId string) {
	cs.hub.AddClient(c, chatId)
}

// RemoveClient removes a client from its current chat room.
func (cs *ChatService) RemoveClient(c domain.Client) {
	cs.hub.RemoveClient(c)
}

// PublishToBroker marshals the payload to JSON and publishes it to the message broker.
func (cs *ChatService) PublishToBroker(ctx context.Context, payload domain.Payload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[chat] publish to broker: marshal error: %v", err)
		return err
	}

	err = cs.publisher.Publish(ctx, body)
	if err != nil {
		log.Printf("[chat] publish to broker: amqp error: %v", err)
		return err
	}

	return nil
}

// HandleIncomingMessage processes a message received from a WebSocket client.
// It adds the client to the appropriate chat room and publishes the message to the broker.
func (cs *ChatService) HandleIncomingMessage(ctx context.Context, c domain.Client, payload domain.Payload) error {
	cs.AddClient(c, payload.ChatID)

	if err := cs.PublishToBroker(ctx, payload); err != nil {
		log.Printf("[chat] handle incoming message: publish to broker error: %v", err)
		return err
	}

	return nil
}

// GetMessagesByChat returns paginated messages for a given chat ID.
func (cs *ChatService) GetMessagesByChat(ctx context.Context, chatID string, limit, offset int) (*domain.GetMessagesResponse, error) {
	msgs, err := cs.repo.GetMessagesByChat(ctx, chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	resp := &domain.GetMessagesResponse{
		ChatID: chatID,
		Limit:  limit,
		Offset: offset,
		Total:  len(msgs),
	}
	for _, m := range msgs {
		resp.Messages = append(resp.Messages, domain.MessageResponse{
			ID:        m.ID,
			UserID:    m.UserID,
			ChatID:    m.ChatID,
			Text:      m.Text,
			CreatedAt: m.CreatedAt.Time,
		})
	}
	return resp, nil
}

// BroadcastMessage subscribes to the Redis PubSub channel and broadcasts
// received messages to all clients in the corresponding chat room.
// This enables cross-instance message delivery.
func (cs *ChatService) BroadcastMessage(ctx context.Context) error {
	out, err := cs.pubsubsubscriber.Subscribe(ctx, config.Redis.PUBSUB_CHANNEL)
	if err != nil {
		log.Printf("[chat] subscribe to redis channel: %v", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-out:
			if !ok {
				return nil
			}
			var payload domain.Payload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("[chat] broadcast message: unmarshal error: %v", err)
				continue
			}

			cs.hub.Broadcast(payload)
		}
	}
}
