package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	AI aiClient // interface để mock test
}

type aiClient interface {
	ClassifyHazardText(ctx context.Context, text string) (urg, itype string, conf float64, err error)
}

type CreateReportRequest struct {
	UserID      string  `json:"userId" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
}

// thêm urgency, incident_type, confidence vào response
func (h *ReportHandler) Create(c *gin.Context) {
	var in CreateReportRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	urg, itype, conf, err := h.AI.ClassifyHazardText(c.Request.Context(), in.Description)
	if err != nil {
		urg, itype, conf = "MEDIUM", "OTHER", 0.0
	}

	// TODO: lưu DB, publish sự kiện nếu cần
	c.JSON(http.StatusCreated, gin.H{
		"ok": true, "ai": gin.H{"urgency": urg, "incident_type": itype, "confidence": conf},
	})
}
