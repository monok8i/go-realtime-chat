package rabbitmq

import (
	"context"

	"go-realtime-chat/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer consumes messages from a RabbitMQ queue with manual acknowledgment.
type Consumer struct {
	channel *amqp.Channel
	queue   amqp.Queue
}

// NewConsumer creates a new consumer connected to the specified queue.
// The queue is declared as durable if it does not exist.
func NewConsumer(conn *amqp.Connection, queueName string) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &Consumer{channel: ch, queue: q}, nil
}

// Consume starts consuming messages from the queue and returns a channel of domain messages.
// Each message includes an Ack function for manual acknowledgment.
func (c *Consumer) Consume(ctx context.Context) (<-chan domain.IncomingBrokerMessage, error) {
	deliveries, err := c.channel.Consume(c.queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	out := make(chan domain.IncomingBrokerMessage)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}
				select {
				case out <- domain.IncomingBrokerMessage{
					Body: d.Body,
					Ack:  func() error { return d.Ack(false) },
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}
