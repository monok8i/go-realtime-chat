package domain

import (
	"context"
)

type Client interface {
	WritePump(ctx context.Context)
	ReadPump(ctx context.Context, onMessage func(ctx context.Context, c Client, payload Payload) error)
	ChatID() string
	SetChatID(id string)
	Send(payload Payload) bool
}

type Hub interface {
	AddClient(client Client, chatId string)
	RemoveClient(client Client)
	Broadcast(chatId string, payload Payload)
}

type ChatService interface {
	AddClient(c Client, chatId string)
	RemoveClient(c Client)
	HandleIncomingMessage(ctx context.Context, c Client, payload Payload) error
	PublishToBroker(ctx context.Context, payload Payload) error
}

type QueuePublisher interface {
	Publish(ctx context.Context, body []byte) error
}

type QueueConsumer interface {
	Consume(ctx context.Context) (<-chan IncomingBrokerMessage, error)
}

type PubSubPublisher interface {
	Publish(ctx context.Context, channel string, body []byte) error
}

type PubSubSubscriber interface {
	Subscribe(ctx context.Context, channel string) (<-chan *IncomingPubSubMessage, error)
}

type MessageRepository interface {
	SaveMessage(ctx context.Context, payload Payload) error
}
