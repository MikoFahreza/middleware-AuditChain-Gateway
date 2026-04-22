package audit

import (
	"github.com/gin-gonic/gin"

	// Sesuaikan "go-blockchain-api" dengan nama module di go.mod Anda jika berbeda
	"go-blockchain-api/internal/middleware"
)

// RegisterRoutes mendaftarkan endpoint khusus untuk membaca data audit
func RegisterRoutes(routerGroup *gin.RouterGroup, h *Handler) {
	// Modul ini memegang kendali penuh atas rute "/dashboard"
	dashAPI := routerGroup.Group("/dashboard")

	// Pasang Satpam (JWT) khusus untuk grup ini
	dashAPI.Use(middleware.JWTAuth())
	{
		dashAPI.GET("/stats", h.GetStats)
		dashAPI.GET("/logs", h.GetRecentLogs)
		dashAPI.GET("/verify/:hash", h.VerifyLog)
		dashAPI.GET("/fabric/:anchor_id", h.GetFabricRecord)
		dashAPI.POST("/verify-data", h.VerifyData)
		dashAPI.GET("/inventory", h.GetResourceInventory)
		dashAPI.GET("/verify-resource/:resource", h.VerifyResourceHistory)
	}
}
