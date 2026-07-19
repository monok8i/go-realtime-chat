package domain

import (
	"context"
)

// Client represents a WebSocket connection with read/write capabilities.
type Client interface {
	WritePump(ctx context.Context)
	ReadPump(ctx context.Context, onMessage func(ctx context.Context, c Client, payload Payload) error)
	ChatID() string
	SetChatID(id string)
	Send(payload Payload) bool
}

// Hub manages WebSocket clients grouped by chat rooms.
type Hub interface {
	AddClient(client Client, chatId string)
	RemoveClient(client Client)
	Broadcast(payload Payload)
}

// ChatService defines business logic for handling chat messages.
type ChatService interface {
	AddClient(c Client, chatId string)
	RemoveClient(c Client)
	HandleIncomingMessage(ctx context.Context, c Client, payload Payload) error
	PublishToBroker(ctx context.Context, payload Payload) error
}

// QueuePublisher sends messages to a message broker queue.
type QueuePublisher interface {
	Publish(ctx context.Context, body []byte) error
}

// QueueConsumer receives messages from a message broker queue.
type QueueConsumer interface {
	Consume(ctx context.Context) (<-chan IncomingBrokerMessage, error)
}

// PubSubPublisher broadcasts messages to a PubSub channel.
type PubSubPublisher interface {
	Publish(ctx context.Context, channel string, body []byte) error
}

// PubSubSubscriber subscribes to a PubSub channel and returns a stream of messages.
type PubSubSubscriber interface {
	Subscribe(ctx context.Context, channel string) (<-chan *IncomingPubSubMessage, error)
}
