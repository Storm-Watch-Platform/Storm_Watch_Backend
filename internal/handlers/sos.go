package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/realtime"

	"github.com/gin-gonic/gin"
)

type SOSHandler struct {
	AI aiSOSClient
	RT realtime.Broadcaster
}

type aiSOSClient interface {
	SosRaise(ctx context.Context, body aiclient.SosRaiseJSONRequestBody) (*aiclient.SosRaiseResponse, error)
}

type SosRaiseRequest struct {
	UserID    string  `json:"userId" binding:"required"`
	AlertBody string  `json:"alert_body" binding:"required"`
	Lat       float64 `json:"lat" binding:"required"`
	Lon       float64 `json:"lon" binding:"required"`
	RadiusM   int     `json:"radius_m" binding:"required"`
	TtlMin    int     `json:"ttl_min" binding:"required"`
}

func (h *SOSHandler) Raise(c *gin.Context) {
	var in SosRaiseRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.AI.SosRaise(c.Request.Context(), aiclient.SosRaiseJSONRequestBody{
		AlertBody: in.AlertBody, Lat: in.Lat, Lon: in.Lon, RadiusM: in.RadiusM, TtlMin: in.TtlMin,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	if h.RT != nil {
		h.RT.BroadcastSOS(resp.SosId, in.Lat, in.Lon, in.RadiusM, resp.ExpiresAt.Format(time.RFC3339), in.AlertBody)
	}
	c.JSON(http.StatusOK, gin.H{
		"ok": true, "sos_id": resp.SosId, "center": []float64{in.Lat, in.Lon},
		"radius_m": in.RadiusM, "expires_at": resp.ExpiresAt,
	})
}
