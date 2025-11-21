package route

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func forwardAI(c *gin.Context, target string) {
	reqBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := http.Post(target, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		c.JSON(503, gin.H{"error": "AI service unavailable"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

func AiProxyRoutes(r *gin.Engine) {
	r.POST("/ai/classify/hazard-text", func(c *gin.Context) {
		forwardAI(c, "http://127.0.0.1:8001/classify/hazard-text")
	})
	r.POST("/ai/presence/update", func(c *gin.Context) {
		forwardAI(c, "http://127.0.0.1:8001/presence/update")
	})
	r.POST("/ai/sos/raise", func(c *gin.Context) {
		forwardAI(c, "http://127.0.0.1:8001/sos/raise")
	})
}
