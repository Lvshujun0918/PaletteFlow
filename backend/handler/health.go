package handler

import (
	"ai-color-palette/logging"

	"github.com/gin-gonic/gin"
)

func HealthHandler(c *gin.Context) {
	logging.Info("health.check", "health check success", logging.Fields{
		"request_id": logging.RequestIDFromGin(c),
	})
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
