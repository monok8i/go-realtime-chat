package rabbitmq

import (
	"context"

	"go-realtime-chat/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewRabbitMQConsumer(conn *amqp.Connection, queueName string) (*RabbitMQConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &RabbitMQConsumer{channel: ch, queue: q}, nil
}

func (c *RabbitMQConsumer) Consume(ctx context.Context) (<-chan domain.IncomingBrokerMessage, error) {
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
