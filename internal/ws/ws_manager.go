package ws

import (
	"encoding/json"
	"sync"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn          *websocket.Conn
	UserID        string
	Subscriptions map[string]bool
}

type WSManager struct {
	mu    sync.RWMutex
	users map[string][]*Client
	temp  []*Client
}

func NewWSManager() *WSManager {
	return &WSManager{
		users: make(map[string][]*Client),
		temp:  []*Client{},
	}
}

func (m *WSManager) AddTempClient(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.temp = append(m.temp, c)
}

func (m *WSManager) PromoteTempClient(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[c.UserID] = append(m.users[c.UserID], c)
	for i, t := range m.temp {
		if t == c {
			m.temp = append(m.temp[:i], m.temp[i+1:]...)
			break
		}
	}
}

func (m *WSManager) RemoveClient(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if list, ok := m.users[c.UserID]; ok {
		newList := []*Client{}
		for _, cc := range list {
			if cc != c {
				newList = append(newList, cc)
			}
		}
		if len(newList) == 0 {
			delete(m.users, c.UserID)
		} else {
			m.users[c.UserID] = newList
		}
	}
	for i, t := range m.temp {
		if t == c {
			m.temp = append(m.temp[:i], m.temp[i+1:]...)
			break
		}
	}
}

func (m *WSManager) Subscribe(c *Client, targetUserID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c.Subscriptions == nil {
		c.Subscriptions = make(map[string]bool)
	}
	c.Subscriptions[targetUserID] = true
}

func (m *WSManager) Unsubscribe(c *Client, targetUserID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c.Subscriptions != nil {
		delete(c.Subscriptions, targetUserID)
	}
}

// ----------------------------
// Broadcast Location object
// ----------------------------
func (m *WSManager) BroadcastLocation(userID string, loc *domain.Location) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.Marshal(loc)
	if err != nil {
		return // nếu marshal lỗi thì bỏ
	}

	for _, clients := range m.users {
		for _, c := range clients {
			if c.Subscriptions != nil && c.Subscriptions[userID] {
				msg := "MESSAGE\ncontent-type:application/json\n\n" + string(data) + "\x00"
				c.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
			}
		}
	}
}

func (w *WSManager) GetConnByUser(userID string) *websocket.Conn {
	w.mu.RLock()
	defer w.mu.RUnlock()

	clients, ok := w.users[userID]
	if !ok || len(clients) == 0 {
		return nil
	}

	// trả về connection đầu tiên tạm thời
	return clients[0].Conn
}

func (w *WSManager) SendToClient(c *Client, destination string, payload interface{}) error {
	data, _ := json.Marshal(payload)
	frame := "SEND\n" +
		"destination:/user/" + c.UserID + "/" + destination + "\n" +
		"content-type:application/json\n\n" +
		string(data) + "\x00"

	return c.Conn.WriteMessage(websocket.TextMessage, []byte(frame))
}

func (m *WSManager) BroadcastSOS(userIDs []string, alert *domain.Alert) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.Marshal(alert)
	if err != nil {
		return
	}

	for _, uid := range userIDs {
		clients, ok := m.users[uid]
		if !ok {
			continue
		}
		for _, c := range clients {
			frame := "SEND\n" +
				"destination:/user/" + c.UserID + "/alert_broadcast\n" +
				"content-type:application/json\n\n" +
				string(data) + "\x00"

			c.Conn.WriteMessage(websocket.TextMessage, []byte(frame))
		}
	}
}
