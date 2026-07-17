// Package ws provides the WebSocket client abstraction and in-memory chat room hub.
package ws

import (
	"go-realtime-chat/internal/domain"
	"log"
	"sync"
)

// Hub manages WebSocket clients grouped by chat rooms.
// It provides thread-safe operations for adding, removing, and broadcasting to clients.
type Hub struct {
	mu    sync.RWMutex
	chats map[string]map[domain.Client]struct{}
}

// NewHub creates a new empty Hub.
func NewHub() *Hub {
	return &Hub{
		chats: make(map[string]map[domain.Client]struct{}),
	}
}

// AddClient registers a client in the given chat room.
// If the client was previously in another room, it is moved.
func (h *Hub) AddClient(c domain.Client, chatId string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if c.ChatID() == chatId {
		return
	}

	if c.ChatID() != "" {
		delete(h.chats[c.ChatID()], c)
	}

	if h.chats[chatId] == nil {
		h.chats[chatId] = make(map[domain.Client]struct{})
	}
	h.chats[chatId][c] = struct{}{}
	c.SetChatID(chatId)
}

// RemoveClient unregisters a client from its current chat room.
func (h *Hub) RemoveClient(c domain.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if c.ChatID() != "" {
		delete(h.chats[c.ChatID()], c)
		if len(h.chats[c.ChatID()]) == 0 {
			delete(h.chats, c.ChatID())
		}
		c.SetChatID("")
	}
}

// Broadcast sends a payload to all clients in the specified chat room.
// Clients with full send buffers are skipped and logged.
func (h *Hub) Broadcast(chatId string, payload domain.Payload) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.chats[chatId] {
		if !client.Send(payload) {
			log.Printf("[hub] broadcast: client buffer full, skipping")
		}
	}
}
