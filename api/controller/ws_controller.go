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

// MAIN WEBSOCKET ENDPOINT (FE connect STOMP vào đây)
func (c *WSController) HandleWS(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		Conn: conn,
	}

	// tạm lưu trước CONNECT frame
	c.WSManager.AddTempClient(client)

	// xử lý các frame STOMP trong goroutine riêng
	go c.handleFrames(client)
}

func (c *WSController) handleFrames(client *ws.Client) {
	defer func() {
		c.WSManager.RemoveClient(client)
		client.Conn.Close()
	}()

	for {
		_, raw, err := client.Conn.ReadMessage()
		if err != nil {
			return
		}

		frame := parseFrame(string(raw))

		switch frame.Command {
		case "CONNECT":
			client.UserID = frame.Headers["user-id"]
			client.IsBackground = frame.Headers["background"] == "true"

			c.WSManager.PromoteTempClient(client)

			resp := "CONNECTED\nversion:1.2\n\n\x00"
			client.Conn.WriteMessage(websocket.TextMessage, []byte(resp))

		case "SUBSCRIBE":
			target := frame.Headers["target-user"]
			c.WSManager.Subscribe(client, target)

		case "UNSUBSCRIBE":
			target := frame.Headers["target-user"]
			c.WSManager.Unsubscribe(client, target)

		case "SEND":
			// gửi location → push tới những client đã subscribe sender
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
