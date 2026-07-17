package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisPubSubPublisher struct {
	Client *redis.Client
}

func NewRedisPubSubPublisher(c *redis.Client) *RedisPubSubPublisher {
	return &RedisPubSubPublisher{Client: c}
}

func (r RedisPubSubPublisher) Publish(ctx context.Context, channel string, body []byte) error {
	if err := r.Client.Publish(ctx, channel, body).Err(); err != nil {
		log.Printf("PubSubPublish: failed to publish message to redis channel: %v", err)
		return err
	}

	return nil
}
