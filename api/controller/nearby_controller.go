// ---------- controller/nearby_controller.go ----------
package controller

import (
	"net/http"
	"strconv"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/internal/ws"
	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/usecase"
	"github.com/gin-gonic/gin"
)

type NearbyController struct {
	AlertUC  *usecase.AlertUseCase
	ReportUC *usecase.ReportUseCase
	WS       *ws.WSManager
}

func NewNearbyController(alertUC *usecase.AlertUseCase, reportUC *usecase.ReportUseCase, ws *ws.WSManager) *NearbyController {
	return &NearbyController{
		AlertUC:  alertUC,
		ReportUC: reportUC,
		WS:       ws,
	}
}

// GET /nearby/sos?lat=...&lon=...&km=...
func (c *NearbyController) NearbySOS(ctx *gin.Context) {
	lat, err1 := strconv.ParseFloat(ctx.Query("lat"), 64)
	lon, err2 := strconv.ParseFloat(ctx.Query("lon"), 64)
	km, err3 := strconv.ParseFloat(ctx.Query("km"), 64)

	if err1 != nil || err2 != nil || err3 != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat/lon/km"})
		return
	}

	alerts, err := c.AlertUC.GetNearbyAlerts(ctx, lat, lon, km)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"alerts": alerts})
}

// GET /nearby/report?lat=...&lon=...&km=...
func (c *NearbyController) NearbyReport(ctx *gin.Context) {
	lat, err1 := strconv.ParseFloat(ctx.Query("lat"), 64)
	lon, err2 := strconv.ParseFloat(ctx.Query("lon"), 64)
	km, err3 := strconv.ParseFloat(ctx.Query("km"), 64)

	if err1 != nil || err2 != nil || err3 != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat/lon/km"})
		return
	}

	reports, err := c.ReportUC.GetNearbyReports(ctx, lat, lon, km)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"reports": reports})
}
