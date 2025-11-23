package controller

import (
	"net/http"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/bootstrap"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/ai"
	"github.com/gin-gonic/gin"
)

type AIController struct {
	APIKey string
}

// Khởi tạo AIController bằng Env
func NewAIController(env *bootstrap.Env) *AIController {
	if env.GeminiAPIKey == "" {
		panic("GEMINI_API_KEY not set in Env")
	}

	return &AIController{
		APIKey: env.GeminiAPIKey,
	}
}

// 📍 POST /ai/analyze
func (ac *AIController) Analyze(c *gin.Context) {
	if ac.APIKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GEMINI_API_KEY not set"})
		return
	}

	var data ai.DisasterData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// build prompt
	prompt, err := ai.BuildDisasterPrompt(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// gọi Gemini API với API key từ controller
	result, err := ai.CallGeminiWithKey(prompt, ac.APIKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", []byte(result))
}
