package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// PubSubPublisher publishes messages to a Redis PubSub channel.
type PubSubPublisher struct {
	client *redis.Client
}

// NewPubSubPublisher creates a new Redis PubSub publisher.
func NewPubSubPublisher(c *redis.Client) *PubSubPublisher {
	return &PubSubPublisher{client: c}
}

// Publish sends a message to the specified Redis PubSub channel.
func (r *PubSubPublisher) Publish(ctx context.Context, channel string, body []byte) error {
	if err := r.client.Publish(ctx, channel, body).Err(); err != nil {
		log.Printf("[redis] publish to channel: %v", err)
		return err
	}

	return nil
}
