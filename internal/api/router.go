package api

import (
	"github.com/gin-gonic/gin"

	"go-blockchain-api/internal/api/dashboard"
	"go-blockchain-api/internal/api/ingestion"
)

// SetupRouter bertugas merakit semua rute URL dan mengembalikan instance server Gin
func SetupRouter(ingestionHandler *ingestion.Handler, dashboardHandler *dashboard.Handler) *gin.Engine {
	// Inisialisasi Gin dengan default middleware (Logger & Recovery)
	router := gin.Default()

	// Nanti, Middleware Global (seperti CORS) bisa ditaruh di sini

	// ==========================================
	// GRUP 1: INGESTION API (Sistem Eksternal)
	// ==========================================
	apiV1 := router.Group("/api/v1")
	{
		// Endpoint ini terbuka untuk menerima log dari sistem lain
		apiV1.POST("/logs", ingestionHandler.ReceiveLog)
	}

	// ==========================================
	// GRUP 2: DASHBOARD API (UI/Frontend)
	// ==========================================
	dashAPI := router.Group("/api/dashboard")
	{
		// Nanti, Middleware JWT akan kita pasang khusus di grup ini

		dashAPI.GET("/stats", dashboardHandler.GetStats)
		dashAPI.GET("/verify/:hash", dashboardHandler.VerifyLog)
		dashAPI.GET("/fabric/:anchor_id", dashboardHandler.GetFabricRecord)
	}

	return router
}
