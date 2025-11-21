package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/classify/hazard-text", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"urgency":       "NONE",
			"incident_type": "NONE",
			"confidence":    0.0,
		})
	})

	r.Run(":8001") // chạy server AI trên port 8001
}
