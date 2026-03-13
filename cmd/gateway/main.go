package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"go-blockchain-api/internal/api"
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

	// fabricSvc is now initialized in main and passed as argument
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		log.Println("⚙️  Background Pipeline Worker berjalan...")
		for range ticker.C {
			hashEngine.ProcessPendingLogs()
			aggEngine.ProcessBatch(5)
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

	// Initialize Fabric Gateway
	fabricSvc, err := blockchain.InitFabricGateway(db)
	if err != nil {
		log.Printf("⚠️ Peringatan: Gagal terhubung ke Fabric Gateway. Anchoring di-bypass.\n")
		fabricSvc = nil
	}

	startPipelineWorker(db)

	// 1. Inisialisasi Repository [BARU]
	auditRepo := repository.NewAuditRepository(db)
	ingestionHandler := &ingestion.Handler{DB: db}
	dashboardHandler := &dashboard.Handler{
		Repo:   auditRepo,
		Fabric: fabricSvc,
	}

	// --- [UPDATE] PANGGIL ROUTER YANG SUDAH DIPISAH ---
	// Kita import dari folder "go-blockchain-api/internal/api"
	router := api.SetupRouter(ingestionHandler, dashboardHandler)

	// Gunakan PORT dari .env
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🚀 AuditChain Gateway API berjalan di port %s...\n", port)
	router.Run(":" + port)
}
