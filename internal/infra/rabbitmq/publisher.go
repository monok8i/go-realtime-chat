package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewPublisher(conn *amqp.Connection, queueName string) (*RabbitMQPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &RabbitMQPublisher{channel: ch, queue: q}, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, body []byte) error {
	return p.channel.PublishWithContext(ctx, "", p.queue.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}
