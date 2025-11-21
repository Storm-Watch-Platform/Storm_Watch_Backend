package handlers

import (
	"context"
	"net/http"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/realtime"
	"github.com/gin-gonic/gin"
)

type PresenceHandler struct {
	AI aiPresenceClient
	RT realtime.Broadcaster
}

type aiPresenceClient interface {
	PresenceUpdate(ctx context.Context, lat, lon, acc *float64, status string) (displayUntil string, err error)
}

type PresenceUpdateRequest struct {
	UserID    string   `json:"userId" binding:"required"`
	Lat       float64  `json:"lat" binding:"required"`
	Lon       float64  `json:"lon" binding:"required"`
	AccuracyM *float64 `json:"accuracy_m"`
	Status    string   `json:"status" binding:"oneof=SAFE CAUTION DANGER UNKNOWN"`
}

func (h *PresenceHandler) Update(c *gin.Context) {
	var in PresenceUpdateRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	displayUntil, err := h.AI.PresenceUpdate(c.Request.Context(), &in.Lat, &in.Lon, in.AccuracyM, in.Status)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	// Broadcast cho room/phạm vi group bạn đang active
	if h.RT != nil {
		h.RT.BroadcastPresence(in.UserID, in.Lat, in.Lon, in.Status, displayUntil)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "display_until": displayUntil})
}
