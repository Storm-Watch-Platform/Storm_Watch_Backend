// ---------- controller/alert_controller.go ----------
package controller

import (
	"net/http"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/gin-gonic/gin"
)

type AlertController struct {
	Repo domain.AlertRepository
}

func NewAlertController(repo domain.AlertRepository) *AlertController {
	return &AlertController{Repo: repo}
}

// POST /alert/mock
func (c *AlertController) MockAlert(ctx *gin.Context) {
	var input domain.Alert
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// userID fake nếu chưa có
	if input.UserID == "" {
		input.UserID = "mock_user_" + time.Now().Format("150405")
	}

	// TTL mặc định
	if input.TTLMin == 0 {
		input.TTLMin = 60
	}

	// Tính ExpiresAt nếu chưa có
	if input.ExpiresAt.IsZero() {
		input.ExpiresAt = time.Now().Add(time.Duration(input.TTLMin) * time.Minute)
	}

	// Visibility mặc định
	if input.Visibility == "" {
		input.Visibility = "PUBLIC"
	}

	// Status mặc định
	if input.Status == "" {
		input.Status = "RAISED"
	}

	// Lưu vào DB
	if err := c.Repo.Create(ctx, &input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "alert": input})
}
