// ---------- controller/ws_controller.go ----------
package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

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
				var body struct {
					Action     string  `json:"action"`     // "raise" hoặc "resolve"
					AlertID    string  `json:"alertId"`    // dùng khi resolve
					Body       string  `json:"body"`       // dùng khi raise
					Lat        float64 `json:"lat"`        // dùng khi raise
					Lon        float64 `json:"lon"`        // dùng khi raise
					RadiusM    float64 `json:"radius_m"`   // dùng khi raise
					TTLMin     int     `json:"ttl_min"`    // dùng khi raise
					Visibility string  `json:"visibility"` // dùng khi raise
				}

				if err := json.Unmarshal([]byte(frame.Body), &body); err != nil {
					println("Cannot unmarshal alert frame:", err.Error())
					break
				}

				switch body.Action {
				case "raise":
					alert := &domain.Alert{
						UserID: client.UserID,
						Body:   body.Body,
						Location: domain.GeoPoint{
							Type:        "Point",
							Coordinates: [2]float64{body.Lon, body.Lat},
						},
						RadiusM:    body.RadiusM,
						TTLMin:     body.TTLMin,
						ExpiresAt:  time.Now().Add(time.Duration(body.TTLMin) * time.Minute),
						Visibility: body.Visibility,
					}
					c.AlertUC.Handle(client, alert)

				case "resolve":
					if body.AlertID == "" {
						println("Missing alertId for resolve")
						break
					}

					err := c.AlertUC.Resolve(client, body.AlertID)
					if err != nil {
						println("Failed to resolve alert:", err.Error())
					}
				}
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
				var body struct {
					Type        string  `json:"type"`
					Detail      string  `json:"detail"`
					Description string  `json:"description"`
					Image       string  `json:"image"`
					Lat         float64 `json:"lat"`
					Lon         float64 `json:"lon"`
					Timestamp   int64   `json:"timestamp"`
				}

				if err := json.Unmarshal([]byte(frame.Body), &body); err != nil {
					break
				}

				report := &domain.Report{
					UserID:      client.UserID,
					Type:        body.Type,
					Detail:      body.Detail,
					Description: body.Description,
					Image:       body.Image,
					Timestamp:   body.Timestamp,
					Location: domain.GeoPoint{
						Type:        "Point",
						Coordinates: [2]float64{body.Lon, body.Lat},
					},
				}

				c.ReportUC.Handle(client, report)
			}
		}
	}
}

type StompFrame struct {
	Command string
	Headers map[string]string
	Body    string
}

// func parseFrame(raw string) StompFrame {
// 	print("PARSE FRAME RAW:", raw)
// 	sc := bufio.NewScanner(strings.NewReader(raw))
// 	f := StompFrame{Headers: map[string]string{}}
// 	if sc.Scan() {
// 		f.Command = sc.Text()
// 	}
// 	for sc.Scan() {
// 		line := sc.Text()
// 		if line == "" {
// 			break
// 		}
// 		kv := strings.SplitN(line, ":", 2)
// 		if len(kv) == 2 {
// 			f.Headers[kv[0]] = kv[1]
// 		}
// 	}
// 	body := ""
// 	for sc.Scan() {
// 		body += sc.Text()
// 	}
// 	f.Body = strings.TrimSuffix(body, "\x00")

// 	print("PARSED FRAME:", body)
// 	return f
// }

func parseFrame(raw string) StompFrame {
	f := StompFrame{Headers: map[string]string{}}
	parts := strings.SplitN(raw, "\n\n", 2)
	headerLines := strings.Split(parts[0], "\n")
	f.Command = headerLines[0]
	for _, line := range headerLines[1:] {
		kv := strings.SplitN(line, ":", 2)
		if len(kv) == 2 {
			f.Headers[kv[0]] = kv[1]
		}
	}
	if len(parts) > 1 {
		f.Body = strings.TrimSuffix(parts[1], "\x00")
	}
	return f
}
