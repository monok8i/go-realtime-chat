package redis

import (
	"context"
	"go-realtime-chat/internal/domain"

	"github.com/redis/go-redis/v9"
)

// RedisPubSubSubscriber subscribes to a Redis PubSub channel and streams messages.
type RedisPubSubSubscriber struct {
	Client *redis.Client
}

// NewRedisPubSubSubscriber creates a new Redis PubSub subscriber.
func NewRedisPubSubSubscriber(c *redis.Client) *RedisPubSubSubscriber {
	return &RedisPubSubSubscriber{Client: c}
}

// Subscribe subscribes to the specified channel and returns a channel of incoming messages.
// The subscription is torn down when the context is cancelled.
func (s *RedisPubSubSubscriber) Subscribe(ctx context.Context, channel string) (<-chan *domain.IncomingPubSubMessage, error) {
	pubsub := s.Client.Subscribe(ctx, channel)

	ch := pubsub.Channel()
	out := make(chan *domain.IncomingPubSubMessage)

	go func() {
		defer pubsub.Close()
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				out <- &domain.IncomingPubSubMessage{Payload: []byte(msg.Payload)}
			}
		}
	}()

	return out, nil
}
