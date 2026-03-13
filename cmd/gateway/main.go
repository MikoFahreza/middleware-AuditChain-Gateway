package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	// Pastikan import ini sesuai dengan nama module di go.mod Anda
	"go-blockchain-api/internal/api"
	"go-blockchain-api/internal/api/dashboard"
	"go-blockchain-api/internal/api/ingestion"
	"go-blockchain-api/internal/blockchain"
	"go-blockchain-api/internal/config"
	"go-blockchain-api/internal/engine/aggregator"
	"go-blockchain-api/internal/engine/hasher"
	"go-blockchain-api/internal/repository"
)

// startPipelineWorker adalah mesin yang berjalan di background setiap 10 detik
func startPipelineWorker(db *gorm.DB, fabricSvc *blockchain.FabricService) {
	hashEngine := &hasher.HasherEngine{DB: db}
	aggEngine := &aggregator.AggregatorEngine{DB: db}

	ticker := time.NewTicker(10 * time.Second)

	go func() {
		log.Println("⚙️ Background Pipeline Worker mulai berjalan...")
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
	// 1. Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		godotenv.Load()
	}

	// 2. Koneksi ke PostgreSQL
	db := config.ConnectDB()

	// 3. Inisialisasi koneksi ke Hyperledger Fabric
	fabricSvc, err := blockchain.InitFabricGateway(db)
	if err != nil {
		log.Println("⚠️ PERINGATAN: Gagal terhubung ke Fabric Gateway!")
		log.Printf("🔍 DETAIL ERROR: %v\n", err)
	}

	// 4. Nyalakan mesin background worker
	startPipelineWorker(db, fabricSvc)

	// 5. Inisialisasi Repository & Handler
	auditRepo := repository.NewAuditRepository(db)
	ingestionHandler := &ingestion.Handler{DB: db}
	dashboardHandler := &dashboard.Handler{
		Repo:   auditRepo,
		Fabric: fabricSvc,
	}

	// 6. Pasang Router yang sudah kita pisahkan ke folder api/
	router := api.SetupRouter(ingestionHandler, dashboardHandler)

	// 7. Jalankan Server API
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🚀 AuditChain Gateway API berjalan di port %s...\n", port)
	router.Run(":" + port)
}
