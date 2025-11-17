package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn         *websocket.Conn
	UserID       string
	IsBackground bool
}

type WSManager struct {
	mu sync.RWMutex

	// userID → list of connections
	users map[string][]*Client

	// B → { A: true }
	followers map[string]map[string]bool

	// temp clients before CONNECT frame
	temp []*Client
}

func NewWSManager() *WSManager {
	return &WSManager{
		users:     map[string][]*Client{},
		followers: map[string]map[string]bool{},
		temp:      []*Client{},
	}
}

// tạm lưu client chưa CONNECT
func (m *WSManager) AddTempClient(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.temp = append(m.temp, c)
}

// chuyển client sau CONNECT
func (m *WSManager) PromoteTempClient(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// add to real
	// thêm conn cho user đó & quăng ra khỏi temp
	m.users[c.UserID] = append(m.users[c.UserID], c)

	// remove from temp
	for i, t := range m.temp {
		if t == c {
			m.temp = append(m.temp[:i], m.temp[i+1:]...)
			break
		}
	}
}

// xoá client bật tab ra ngoài
// tắt mạng, đóng tab -> remove
func (m *WSManager) RemoveClient(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// xóa khỏi users
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

	// xóa khỏi temp
	for i, t := range m.temp {
		if t == c {
			m.temp = append(m.temp[:i], m.temp[i+1:]...)
			break
		}
	}
}

// A follow B
func (m *WSManager) Follow(follower, target string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.followers[target]; !ok {
		m.followers[target] = map[string]bool{}
	}

	m.followers[target][follower] = true
}

// A unfollow B
func (m *WSManager) Unfollow(follower, target string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if set, ok := m.followers[target]; ok {
		delete(set, follower)
		if len(set) == 0 {
			delete(m.followers, target)
		}
	}
}

// B gửi vị trí → push đến followers
func (m *WSManager) SendLocationFromUser(userID string, body string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	follows, ok := m.followers[userID]
	if !ok {
		return // không có ai theo dõi
	}

	for followerID := range follows {
		list := m.users[followerID]
		for _, cli := range list {
			// gửi STOMP MESSAGE frame
			msg := "MESSAGE\n" +
				"content-type:application/json\n\n" +
				body + "\x00"

			cli.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}
}
