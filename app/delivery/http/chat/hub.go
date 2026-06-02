package http_chat

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Hub manages active WebSocket clients and broadcasts messages to them.
type Hub struct {
	clients    map[string]map[*Client]bool // userID -> set of connections
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// Client represents a single WebSocket connection.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID string
	send   chan []byte
}

// WSMessage is the envelope sent over WebSocket to the client.
type WSMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's event loop. Call this as a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()
			logrus.Infof("WebSocket client registered: userID=%s", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.clients[client.userID]; ok {
				if _, exists := conns[client]; exists {
					delete(conns, client)
					close(client.send)
					if len(conns) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}
			h.mu.Unlock()
			logrus.Infof("WebSocket client unregistered: userID=%s", client.userID)
		}
	}
}

// SendToUser sends a message to all active connections for a given user.
func (h *Hub) SendToUser(userID string, event string, payload interface{}) {
	msg := WSMessage{
		Event: event,
		Data:  payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		logrus.Errorf("WebSocket marshal error: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if conns, ok := h.clients[userID]; ok {
		for client := range conns {
			select {
			case client.send <- data:
			default:
				// client too slow, drop
				close(client.send)
				delete(conns, client)
			}
		}
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub.
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
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("WebSocket read error: %v", err)
			}
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(payload, &msg); err == nil {
			if msg.Event == "typing" {
				if dataMap, ok := msg.Data.(map[string]interface{}); ok {
					if targetUserID, ok := dataMap["target_user_id"].(string); ok {
						dataMap["sender_user_id"] = c.userID
						msg.Data = dataMap
						c.hub.SendToUser(targetUserID, "typing", dataMap)
					}
				}
			}
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
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
