package ingestion

import (
	"go-blockchain-api/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(routerGroup *gin.RouterGroup, h *Handler) {
	// Modul ini membuat rute untuk "/logs"
	logsRoutes := routerGroup.Group("/logs")

	// 👇 PASANG MIDDLEWARE DI SINI
	logsRoutes.Use(middleware.APIKeyAuth(&gorm.DB{}))
	{
		// Rute ini sekarang terlindungi oleh API Key
		logsRoutes.POST("/", h.ReceiveLog)
	}
}
