// Package domain defines the core domain types and interfaces for the chat system.
package domain

// Payload represents a chat message exchanged between clients through the system.
type Payload struct {
	UserID  int32  `json:"user_id"`
	ChatID  string `json:"chat_id"`
	Message string `json:"message"`
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
