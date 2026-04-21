package ingestion

import (
	"go-blockchain-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, h *Handler) {
	logsRoutes := routerGroup.Group("/logs")
	logsRoutes.Use(middleware.APIKeyAuth(h.DB))
	{
		// Rute ini sekarang terlindungi oleh API Key
		logsRoutes.POST("/", h.ReceiveLog)
	}
}
