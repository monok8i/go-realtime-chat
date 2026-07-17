package ws

import (
	"context"
	"encoding/json"
	"go-realtime-chat/internal/domain"
	"log"

	"github.com/gorilla/websocket"
)

// Client represents a single WebSocket connection with read/write capabilities.
type Client struct {
	conn   *websocket.Conn
	send   chan domain.Payload
	chatId string
}

// NewClient creates a new Client wrapping the given WebSocket connection.
// The send buffer is initialised with capacity 16.
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan domain.Payload, 16),
	}
}

// WritePump reads messages from the send channel and writes them to the WebSocket connection.
// It exits when the context is cancelled or the send channel is closed.
func (c *Client) WritePump(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("[ws] writePump: marshal error: %v", err)
				continue
			}
			err = c.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("[ws] writePump: write error: %v", err)
				return
			}
		}
	}
}

// ChatID returns the ID of the chat room the client is currently in.
func (c *Client) ChatID() string {
	return c.chatId
}

// SetChatID sets the ID of the chat room the client is assigned to.
func (c *Client) SetChatID(id string) {
	c.chatId = id
}

// Send attempts to enqueue a payload for writing. Returns false if the buffer is full.
func (c *Client) Send(payload domain.Payload) bool {
	select {
	case c.send <- payload:
		return true
	default:
		return false
	}
}

// ReadPump reads messages from the WebSocket connection and passes them to onMessage.
// It closes the connection when the read loop exits.
func (c *Client) ReadPump(ctx context.Context, onMessage func(ctx context.Context, cl domain.Client, payload domain.Payload) error) {
	defer func() {
		if err := c.conn.Close(); err != nil {
			log.Printf("[ws] close websocket connection: %v", err)
		}
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("[ws] readPump: read error: %v", err)
			return
		}

		var payload domain.Payload
		if err := json.Unmarshal(data, &payload); err != nil {
			log.Printf("[ws] readPump: unmarshal error: %v", err)
			continue
		}

		if err := onMessage(ctx, c, payload); err != nil {
			log.Printf("[ws] readPump: onMessage error: %v", err)
			continue
		}
	}
}
