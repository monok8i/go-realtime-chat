// Package domain defines the core domain types and interfaces for the chat system.
package domain

import (
	"time"
)

// Payload represents a chat message exchanged between clients through the system.
type Payload struct {
	UserID  int32  `json:"user_id"`
	ChatID  string `json:"chat_id"`
	Message string `json:"message"`
}

// Message represents a chat message received from database.
type Message struct {
	ID        int64
	UserID    int32
	ChatID    string
	Text      string
	CreatedAt time.Time
}

// MessageResponse represents a single message returned by the API.
type MessageResponse struct {
	ID        int64     `json:"id"`
	UserID    int32     `json:"user_id"`
	ChatID    string    `json:"chat_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// GetMessagesResponse is the paginated response for the messages endpoint.
type GetMessagesResponse struct {
	ChatID   string            `json:"chat_id"`
	Messages []MessageResponse `json:"messages"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	Total    int               `json:"total"`
}

// IncomingBrokerMessage wraps a message received from the message broker (RabbitMQ).
// Body contains the raw message bytes, Ack provides a way to acknowledge processing.
type IncomingBrokerMessage struct {
	Body []byte
	Ack  func() error
}

// IncomingPubSubMessage wraps a message received from the PubSub system (Redis).
type IncomingPubSubMessage struct {
	Payload []byte
}
