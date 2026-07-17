package redis

import (
	"context"
	"go-realtime-chat/internal/domain"

	"github.com/redis/go-redis/v9"
)

type RedisPubSubSubscriber struct {
	Client *redis.Client
}

func NewRedisPubSubSubscriber(c *redis.Client) *RedisPubSubSubscriber {
	return &RedisPubSubSubscriber{Client: c}
}

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
