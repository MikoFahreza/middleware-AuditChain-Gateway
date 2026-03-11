package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"go-blockchain-api/internal/api/dashboard"
	"go-blockchain-api/internal/api/ingestion"
	"go-blockchain-api/internal/blockchain"
	"go-blockchain-api/internal/config"
	"go-blockchain-api/internal/engine/aggregator"
	"go-blockchain-api/internal/engine/hasher"
	"go-blockchain-api/internal/repository"
)

func startPipelineWorker(db *gorm.DB) {
	hashEngine := &hasher.HasherEngine{DB: db}
	aggEngine := &aggregator.AggregatorEngine{DB: db}

	fabricSvc, err := blockchain.InitFabricGateway(db)
	if err != nil {
		log.Printf("⚠️ Peringatan: Gagal terhubung ke Fabric Gateway. Anchoring di-bypass.\n")
	}

	ticker := time.NewTicker(10 * time.Second)

	go func() {
		log.Println("⚙️  Background Pipeline Worker berjalan...")
		for range ticker.C {
			hashEngine.ProcessPendingLogs()
			aggEngine.ProcessBatch(5)
			if fabricSvc != nil {
				fabricSvc.AnchorPendingRoots()
			}
		}
	}()
}

func main() {
	// Karena main.go sekarang ada di cmd/gateway/, kita harus pastikan ia bisa baca .env di root folder
	if err := godotenv.Load("../../.env"); err != nil {
		// Fallback mencari di direktori saat command dijalankan
		godotenv.Load()
	}

	db := config.ConnectDB()
	startPipelineWorker(db)

	// 1. Inisialisasi Repository [BARU]
	auditRepo := repository.NewAuditRepository(db)

	// 2. Inject Repository ke Handler Dashboard (Ingestion sementara tetap pakai DB langsung atau bisa diubah nanti)
	ingestionHandler := &ingestion.Handler{DB: db}
	dashboardHandler := &dashboard.Handler{Repo: auditRepo} // [UPDATE]

	router := gin.Default()

	// --- ROUTING GRUP INGESTION ---
	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/logs", ingestionHandler.ReceiveLog)
	}

	// --- ROUTING GRUP DASHBOARD ---
	dashAPI := router.Group("/api/dashboard")
	{
		dashAPI.GET("/stats", dashboardHandler.GetStats)
		// Endpoint Baru untuk Audit Verifikasi [BARU]
		dashAPI.GET("/verify/:hash", dashboardHandler.VerifyLog)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 AuditChain Gateway berjalan di port %s...\n", port)
	router.Run(":" + port)
}
