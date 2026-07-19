package postgres

import (
	"context"
	"go-realtime-chat/internal/domain"
	"go-realtime-chat/internal/infra/postgres/gen"
)

type MessageRepository struct {
	q *gen.Queries
}

func NewMessageRepository(q *gen.Queries) *MessageRepository {
	return &MessageRepository{q: q}
}

func (r *MessageRepository) CreateNewMessage(ctx context.Context, payload domain.Payload) error {
	_, err := r.q.CreateMessage(ctx, gen.CreateMessageParams{
		UserID: int32(payload.UserID),
		ChatID: payload.ChatID,
		Text:   payload.Message,
	})
	return err
}

func (r *MessageRepository) GetMessagesByChat(ctx context.Context, chatID string) ([]domain.Payload, error) {
	msgs, err := r.q.GetMessagesByChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Payload, len(msgs))
	for i, m := range msgs {
		result[i] = domain.Payload{
			UserID:  int(m.UserID),
			ChatID:  m.ChatID,
			Message: m.Text,
		}
	}
	return result, nil
}
