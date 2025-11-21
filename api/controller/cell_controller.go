package controller

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
	"github.com/gin-gonic/gin"
)

type CellController struct {
	CellUsecase domain.CellUsecase
}

// ----------------------
// Request structs
// ----------------------
type UpdateCellRequest struct {
	Lat       float64 `json:"lat" binding:"required"`
	Lon       float64 `json:"lon" binding:"required"`
	RiskScore float64 `json:"riskScore" binding:"required"`
	Label     string  `json:"label" binding:"required"`
}

type UpdateCellsRadiusRequest struct {
	Lat     float64 `json:"lat" binding:"required"`
	Lon     float64 `json:"lon" binding:"required"`
	Radius  float64 `json:"radius" binding:"required"`
	RiskInc float64 `json:"riskInc" binding:"required"` // tăng/giảm risk
	Label   string  `json:"label"`
}

// ----------------------
// Helpers
// ----------------------

// snapGrid chuẩn hóa 1 điểm thành cell ~100x100m
func snapGrid(lat, lon float64) (float64, float64) {
	const cellSize = 0.001
	latNorm := math.Floor(lat/cellSize) * cellSize
	lonNorm := math.Floor(lon/cellSize) * cellSize
	return latNorm, lonNorm
}

// ----------------------
// Handlers
// ----------------------

// GetCellByLatLon fetch a single cell at given lat/lon
func (cc *CellController) GetCellByLatLon(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lon, err2 := strconv.ParseFloat(lonStr, 64)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat/lon"})
		return
	}

	cell, err := cc.CellUsecase.FetchByLatLon(c, lat, lon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Nếu chưa có cell, trả về mặc định risk=0
	if cell == nil {
		latSnap, lonSnap := snapGrid(lat, lon)
		cell = &domain.Cell{
			Center:    [2]float64{lonSnap, latSnap}, // [lon, lat]
			RiskScore: 0,
			Label:     "SAFE",
			UpdatedAt: time.Now(),
		}
	}

	c.JSON(http.StatusOK, cell)
}

// GetCellsByRadius fetch cells within radius
func (cc *CellController) GetCellsByRadius(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	radiusStr := c.Query("radius")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Invalid lat"})
		return
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Invalid lon"})
		return
	}
	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "Invalid radius"})
		return
	}

	cells, err := cc.CellUsecase.GetCellsByRadius(c, lat, lon, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, cells)
}

// UpdateCell create or update a single cell
func (cc *CellController) UpdateCell(c *gin.Context) {
	var req UpdateCellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	latSnap, lonSnap := snapGrid(req.Lat, req.Lon)
	cell := &domain.Cell{
		Center:    [2]float64{lonSnap, latSnap}, // lưu [lon, lat] để index
		RiskScore: req.RiskScore,
		Label:     req.Label,
		UpdatedAt: time.Now(),
	}

	if err := cc.CellUsecase.Upsert(c, cell); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, cell)
}

// UpdateCellsByRadius create/update cells within radius, increasing/decreasing risk
func (cc *CellController) UpdateCellsByRadius(c *gin.Context) {
	var req UpdateCellsRadiusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: err.Error()})
		return
	}

	updated, inserted, err := cc.CellUsecase.UpdateCellsByRadius(c, req.Lat, req.Lon, req.Radius, req.RiskInc, req.Label)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"updated":  updated,
		"inserted": inserted,
	})
}
