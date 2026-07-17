package ws

import (
	"go-realtime-chat/internal/domain"
	"log"
	"sync"
)

type Hub struct {
	mu    sync.RWMutex
	chats map[string]map[domain.Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		chats: make(map[string]map[domain.Client]struct{}),
	}
}

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

func (h *Hub) Broadcast(chatId string, payload domain.Payload) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.chats[chatId] {
		if !client.Send(payload) {
			log.Printf("broadcast: client buffer full, skipping")
		}
	}
}
