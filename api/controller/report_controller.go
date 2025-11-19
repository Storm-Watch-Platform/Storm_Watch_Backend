package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/gin-gonic/gin"
)

type ReportController struct {
	ReportRepo domain.ReportRepository // dùng interface ReportRepository
	Timeout    time.Duration
}

// POST /report/mock
func (c *ReportController) MockReport(ctx *gin.Context) {
	var input domain.Report
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.UserID == "" {
		input.UserID = "mock_user_" + time.Now().Format("150405")
	}
	if input.Timestamp == 0 {
		input.Timestamp = time.Now().Unix()
	}
	if input.Status == "" {
		input.Status = "RAISED"
	}

	// lưu trực tiếp bằng repository
	ctxDB, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	if err := c.ReportRepo.Create(ctxDB, &input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "report": input})
}
