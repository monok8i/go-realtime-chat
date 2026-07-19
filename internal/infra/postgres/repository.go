// Package postgres provides PostgreSQL-based implementations of domain interfaces.
package postgres

import (
	"context"
	"go-realtime-chat/internal/domain"
	"go-realtime-chat/internal/infra/postgres/gen"
)

// MessageRepository implements the domain.MessageRepository interface using a PostgreSQL database.
type MessageRepository struct {
	q *gen.Queries
}

// NewMessageRepository creates a new MessageRepository backed by the given sqlc queries.
func NewMessageRepository(q *gen.Queries) *MessageRepository {
	return &MessageRepository{q: q}
}

// CreateNewMessage inserts a new chat message into the database.
func (r *MessageRepository) CreateNewMessage(ctx context.Context, payload domain.Payload) error {
	_, err := r.q.CreateMessage(ctx, gen.CreateMessageParams{
		UserID: payload.UserID,
		ChatID: payload.ChatID,
		Text:   payload.Message,
	})
	return err
}

// GetMessagesByChat retrieves all messages for a given chat ID.
func (r *MessageRepository) GetMessagesByChat(ctx context.Context, chatID string) ([]domain.Payload, error) {
	msgs, err := r.q.GetMessagesByChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Payload, len(msgs))
	for i, m := range msgs {
		result[i] = domain.Payload{
			UserID:  m.UserID,
			ChatID:  m.ChatID,
			Message: m.Text,
		}
	}
	return result, nil
}
