package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher publishes messages to a RabbitMQ queue.
type Publisher struct {
	channel *amqp.Channel
	queue   amqp.Queue
}

// NewPublisher creates a new publisher connected to the specified queue.
// The queue is declared as durable if it does not exist.
func NewPublisher(conn *amqp.Connection, queueName string) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &Publisher{channel: ch, queue: q}, nil
}

// Publish sends a message to the configured queue with content type application/json.
func (p *Publisher) Publish(ctx context.Context, body []byte) error {
	return p.channel.PublishWithContext(ctx, "", p.queue.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}
