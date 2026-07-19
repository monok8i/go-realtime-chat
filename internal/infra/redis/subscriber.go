package redis

import (
	"context"
	"go-realtime-chat/internal/domain"

	"github.com/redis/go-redis/v9"
)

// PubSubSubscriber subscribes to a Redis PubSub channel and streams messages.
type PubSubSubscriber struct {
	client *redis.Client
}

// NewPubSubSubscriber creates a new Redis PubSub subscriber.
func NewPubSubSubscriber(c *redis.Client) *PubSubSubscriber {
	return &PubSubSubscriber{client: c}
}

// Subscribe subscribes to the specified channel and returns a channel of incoming messages.
// The subscription is torn down when the context is cancelled.
func (s *PubSubSubscriber) Subscribe(ctx context.Context, channel string) (<-chan *domain.IncomingPubSubMessage, error) {
	pubsub := s.client.Subscribe(ctx, channel)

	ch := pubsub.Channel()
	out := make(chan *domain.IncomingPubSubMessage)

	go func() {
		defer func() { _ = pubsub.Close() }()
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				select {
				case out <- &domain.IncomingPubSubMessage{Payload: []byte(msg.Payload)}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}
