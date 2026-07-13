package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
	rooms  []string
	mu     sync.Mutex
}

func NewClient(hub *Hub, conn *websocket.Conn, userID string, rooms []string) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		rooms:  rooms,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[ws] read error: %v", err)
			}
			break
		}

		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			continue
		}

		switch clientMsg.Type {
		case "subscribe":
			c.hub.mu.Lock()
			c.rooms = append(c.rooms, clientMsg.Room)
			if c.hub.rooms[clientMsg.Room] == nil {
				c.hub.rooms[clientMsg.Room] = make(map[*Client]bool)
			}
			c.hub.rooms[clientMsg.Room][c] = true
			c.hub.mu.Unlock()

		case "unsubscribe":
			c.hub.mu.Lock()
			for i, room := range c.rooms {
				if room == clientMsg.Room {
					c.rooms = append(c.rooms[:i], c.rooms[i+1:]...)
					break
				}
			}
			if clients, ok := c.hub.rooms[clientMsg.Room]; ok {
				delete(clients, c)
			}
			c.hub.mu.Unlock()
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
