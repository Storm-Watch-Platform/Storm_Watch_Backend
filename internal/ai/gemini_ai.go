package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

// -------------------- DATA STRUCT --------------------
type DisasterData struct {
	Alerts  []Alert  `json:"alerts"`
	Reports []Report `json:"reports"`
}

type Alert struct {
	Location    LatLng `json:"location"`
	Description string `json:"description"`
	Urgency     string `json:"urgency"`
	Timestamp   any    `json:"timestamp"`
}

type Report struct {
	Location    LatLng `json:"location"`
	Description string `json:"description"`
	Urgency     string `json:"urgency"`
	Timestamp   any    `json:"timestamp"`
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// -------------------- BUILD PROMPT --------------------
func BuildDisasterPrompt(data DisasterData) (string, error) {
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`
Bạn là trợ lý phân tích thiên tai. Nhiệm vụ:
1. Xác định nguyên nhân chính.
2. Mô tả hiện trạng.
3. Đưa ra khuyến nghị nên làm gì tiếp theo.
4. Tổng hợp mức độ nghiêm trọng.

Trả về **chỉ JSON**:
{
  "cause": "nguyên nhân chính",
  "current_status": "tình hình hiện tại",
  "recommendation": "nên làm gì tiếp theo",
  "severity": "HIGH | MEDIUM | LOW"
}

Dữ liệu:
%s
`, string(raw))

	return prompt, nil
}

// -------------------- CALL GEMINI API --------------------
func CallGeminiWithKey(prompt string, apiKey string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return "", err
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", err
	}

	return result.Text(), nil
}
