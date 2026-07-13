package ws

import (
	"log"
	"sync"
)

type Hub struct {
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	done       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		done:       make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			for _, room := range client.rooms {
				if h.rooms[room] == nil {
					h.rooms[room] = make(map[*Client]bool)
				}
				h.rooms[room][client] = true
			}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			for _, room := range client.rooms {
				if clients, ok := h.rooms[room]; ok {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.rooms, room)
					}
				}
			}
			close(client.send)
			h.mu.Unlock()

		case <-h.done:
			return
		}
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

func (h *Hub) BroadcastToRoom(roomID string, msg []byte) {
	h.mu.RLock()
	clients := h.rooms[roomID]
	h.mu.RUnlock()

	if clients == nil {
		return
	}

	for client := range clients {
		select {
		case client.send <- msg:
		default:
			h.mu.Lock()
			for _, room := range client.rooms {
				if clients, ok := h.rooms[room]; ok {
					delete(clients, client)
				}
			}
			close(client.send)
			h.mu.Unlock()
		}
	}
}

func (h *Hub) CreateRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
}

func (h *Hub) DeleteRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms, roomID)
}

func (h *Hub) Stop() {
	close(h.done)
}

func (h *Hub) RoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	seen := make(map[*Client]bool)
	for _, clients := range h.rooms {
		for c := range clients {
			seen[c] = true
		}
	}
	return len(seen)
}

func init() {
	log.Println("[ws] package loaded — Hub, Client, Message types ready")
}
