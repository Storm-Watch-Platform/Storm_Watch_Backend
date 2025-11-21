package controller

import (
	"fmt"
	"net/http"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/gin-gonic/gin"
)

type ZoneController struct {
	ZoneUsecase domain.ZoneUsecase
}

// =======================
// Request models
// =======================
type CreateZoneRequest struct {
	Lat       float64 `json:"lat" binding:"required"`
	Lon       float64 `json:"lon" binding:"required"`
	Radius    float64 `json:"radius" binding:"required"`
	RiskScore float64 `json:"riskScore"`
	Label     string  `json:"label"`
}

// =======================
// POST /zones — create 1 zone
// =======================
func (zc *ZoneController) Create(c *gin.Context) {
	var req CreateZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	zone := &domain.Zone{
		Center:    [2]float64{req.Lon, req.Lat},
		Radius:    req.Radius,
		RiskScore: req.RiskScore,
		Label:     req.Label,
	}

	if err := zc.ZoneUsecase.Create(c, zone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, zone)
}

// =======================
// GET /zones?minLat&minLon&maxLat&maxLon
// Lấy zone trong bounding box
// =======================
func (zc *ZoneController) FetchInBounds(c *gin.Context) {
	minLat, ok1 := getFloatQuery(c, "minLat")
	minLon, ok2 := getFloatQuery(c, "minLon")
	maxLat, ok3 := getFloatQuery(c, "maxLat")
	maxLon, ok4 := getFloatQuery(c, "maxLon")

	if !(ok1 && ok2 && ok3 && ok4) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid bounding box"})
		return
	}

	zones, err := zc.ZoneUsecase.FetchInBounds(c, minLat, minLon, maxLat, maxLon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, zones)
}

// =======================
// GET /zones/by-location?lat&lon
// Lấy zone chứa 1 điểm
// =======================
func (zc *ZoneController) FetchByLatLon(c *gin.Context) {
	lat, ok1 := getFloatQuery(c, "lat")
	lon, ok2 := getFloatQuery(c, "lon")
	if !(ok1 && ok2) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid lat/lon"})
		return
	}

	zones, err := zc.ZoneUsecase.FetchAllByLatLon(c, lat, lon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if len(zones) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "no zone found"})
		return
	}

	c.JSON(http.StatusOK, zones)
}

// =======================
// GET /zones/all
// =======================
func (zc *ZoneController) FetchAll(c *gin.Context) {
	zones, err := zc.ZoneUsecase.FetchAll(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, zones)
}

// =======================
// Helper — get float query
// =======================
func getFloatQuery(c *gin.Context, key string) (float64, bool) {
	valStr := c.Query(key)
	if valStr == "" {
		return 0, false
	}
	var val float64
	_, err := fmt.Sscanf(valStr, "%f", &val)
	if err != nil {
		return 0, false
	}
	return val, true
}
