// ---------- controller/ws_controller.go ----------
package controller

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/ws"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/usecase"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSController struct {
	WSManager  *ws.WSManager
	LocationUC *usecase.LocationUseCase
	ReportUC   *usecase.ReportUseCase
	AlertUC    *usecase.AlertUseCase // pointer
}

func (c *WSController) HandleWS(ctx *gin.Context) {
	// nâng cấp kết nối HTTP lên WebSocket
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	// tạo temp client và thêm vào WSManager
	client := &ws.Client{Conn: conn}
	c.WSManager.AddTempClient(client)

	// xử lý các frame từ client
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
		// yêu cầu connect -> tạo userid thật, từ temp client thành chính thức
		case "CONNECT":
			client.UserID = frame.Headers["user-id"]
			c.WSManager.PromoteTempClient(client)
			resp := "CONNECTED\nversion:1.2\n\n\x00"
			client.Conn.WriteMessage(websocket.TextMessage, []byte(resp))

		// chọn client muốn nhận thông báo
		case "SUBSCRIBE":
			target := frame.Headers["target-user"]
			c.WSManager.Subscribe(client, target)

		// hủy muốn nhận thông báo từ client id nào đó
		case "UNSUBSCRIBE":
			target := frame.Headers["target-user"]
			c.WSManager.Unsubscribe(client, target)

		// gửi
		case "SEND":
			msgType := frame.Headers["type"]
			switch msgType {
			case "alert":
				var alert domain.Alert
				err := json.Unmarshal([]byte(frame.Body), &alert)
				if err != nil {
					// handle lỗi
					break
				}
				c.AlertUC.Handle(client.UserID, &alert)
			case "location":
				var body struct {
					Lat       float64 `json:"Lat"`
					Lon       float64 `json:"Lon"`
					AccuracyM float64 `json:"AccuracyM"`
					Status    string  `json:"Status"`
					UpdatedAt int64   `json:"UpdatedAt"`
				}

				if err := json.Unmarshal([]byte(frame.Body), &body); err != nil {
					break
				}

				loc := &domain.Location{
					ID:        client.UserID,
					AccuracyM: body.AccuracyM,
					Status:    body.Status,
					UpdatedAt: body.UpdatedAt / 1000, // ms -> s nếu muốn
					Location: domain.GeoPoint{
						Type:        "Point",
						Coordinates: [2]float64{body.Lon, body.Lat},
					},
				}

				c.LocationUC.Handle(client.UserID, loc)

			case "report":
				var report domain.Report
				err := json.Unmarshal([]byte(frame.Body), &report)
				if err != nil {
					break
				}
				c.ReportUC.Handle(client.UserID, &report)

			}
		}
	}
}

type StompFrame struct {
	Command string
	Headers map[string]string
	Body    string
}

func parseFrame(raw string) StompFrame {
	sc := bufio.NewScanner(strings.NewReader(raw))
	f := StompFrame{Headers: map[string]string{}}
	if sc.Scan() {
		f.Command = sc.Text()
	}
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
	body := ""
	for sc.Scan() {
		body += sc.Text()
	}
	f.Body = strings.TrimSuffix(body, "\x00")
	return f
}
