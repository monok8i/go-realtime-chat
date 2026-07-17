package ws

import (
	"context"
	"encoding/json"
	"go-realtime-chat/internal/domain"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn   *websocket.Conn
	send   chan domain.Payload
	chatId string
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan domain.Payload, 16),
	}
}

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
				log.Printf("writePump: marshal error: %v", err)
				continue
			}
			err = c.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("writePump: write error: %v", err)
				return
			}
		}
	}
}

func (c *Client) ChatID() string {
	return c.chatId
}

func (c *Client) SetChatID(id string) {
	c.chatId = id
}

func (c *Client) Send(payload domain.Payload) bool {
	select {
	case c.send <- payload:
		return true
	default:
		return false
	}
}

func (c *Client) ReadPump(ctx context.Context, onMessage func(ctx context.Context, cl domain.Client, payload domain.Payload) error) {
	defer func() {
		if err := c.conn.Close(); err != nil {
			log.Printf("Failed to close websocket connection: %v", err)
		}
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("readPump: read error: %v", err)
			return
		}

		var payload domain.Payload
		if err := json.Unmarshal(data, &payload); err != nil {
			log.Printf("readPump: unmarshal error: %v", err)
			continue
		}

		onMessage(ctx, c, payload)
	}
}
