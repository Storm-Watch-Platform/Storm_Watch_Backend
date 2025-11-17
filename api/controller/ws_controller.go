package controller

import (
	"bufio"
	"net/http"
	"strings"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// nâng http -> websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSController struct {
	WSManager *ws.WSManager
}

// FE mở socket qua /ws
// MAIN WEBSOCKET ENDPOINT (FE connect stomp vào đây)
func (c *WSController) HandleWS(ctx *gin.Context) {
	// upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	// chưa có user id
	client := &ws.Client{
		Conn: conn,
	}

	// mở xong còn cần handshake nữa, chờ CONNECT frame
	c.WSManager.AddTempClient(client)

	// listen STOMP frames
	// thêm thread để server xử lý tiếp
	go c.handleFrames(client)
}

func (c *WSController) handleFrames(client *ws.Client) {
	// ngắt kết nối hoặc lỗi thì bái bai client
	defer func() {
		c.WSManager.RemoveClient(client)
		client.Conn.Close()
	}()

	for {
		// chờ frame từ client
		_, raw, err := client.Conn.ReadMessage()
		if err != nil {
			return
		}

		frame := parseFrame(string(raw))

		switch frame.Command {

		// xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
		// CONNECT
		// user-id:abc123
		// background:false
		case "CONNECT":
			// cập nhật temp client thành chính thức
			client.UserID = frame.Headers["user-id"]
			client.IsBackground = frame.Headers["background"] == "true"

			c.WSManager.PromoteTempClient(client)

			// reply CONNECTED
			resp := "CONNECTED\nversion:1.2\n\n\x00"
			client.Conn.WriteMessage(websocket.TextMessage, []byte(resp))

		// xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
		case "SUBSCRIBE":
			// follower: A; target-user: B
			// followers[B] = [A1, A2, ...]
			follower := client.UserID // mình follow người ta
			target := frame.Headers["target-user"]
			c.WSManager.Follow(follower, target)

		// xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
		case "UNSUBSCRIBE":
			// bỏ A khỏi followers[B]
			follower := client.UserID
			target := frame.Headers["target-user"]
			c.WSManager.Unfollow(follower, target)

		// xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
		case "SEND":
			// A gửi location -> tìm ai đang follow A để push location
			// SEND location JSON
			c.WSManager.SendLocationFromUser(client.UserID, frame.Body)
		}
	}
}

// SIMPLE STOMP FRAME PARSER
type StompFrame struct {
	Command string
	Headers map[string]string
	Body    string
}

// SEND
// destination:/user/123/location
// content-type:application/json

// {"lat":10,"lng":20}\x00

// input CONNECT\nuser-id:123\nbackground:false\n\n\x00

// StompFrame{
//     Command: "CONNECT",
//     Headers: map[string]string{
//         "user-id": "123",
//         "background": "false",
//     },
//     Body: "",
// }

func parseFrame(raw string) StompFrame {
	sc := bufio.NewScanner(strings.NewReader(raw))
	f := StompFrame{Headers: map[string]string{}}

	// COMMAND
	if sc.Scan() {
		f.Command = sc.Text()
	}

	// HEADERS
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			break
		}
		kv := strings.SplitN(line, ":", 2)
		if len(kv) == 2 {
			f.Headers[kv[0]] = kv[1]
		}
	}

	// BODY
	body := ""
	for sc.Scan() {
		body += sc.Text()
	}

	f.Body = strings.TrimSuffix(body, "\x00")
	return f
}
