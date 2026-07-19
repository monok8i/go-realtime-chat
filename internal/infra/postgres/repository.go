// Package postgres provides PostgreSQL-based implementations of domain interfaces.
package postgres

import (
	"context"
	"math"

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

// GetMessagesByChat retrieves messages for a given chat ID with pagination.
func (r *MessageRepository) GetMessagesByChat(ctx context.Context, chatID string, limit, offset int) ([]gen.Message, error) {
	if limit > math.MaxInt32 {
		limit = math.MaxInt32
	}
	if offset > math.MaxInt32 {
		offset = math.MaxInt32
	}

	return r.q.GetMessagesByChat(ctx, gen.GetMessagesByChatParams{
		ChatID: chatID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
}
