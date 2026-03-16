package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	// Pastikan import ini sesuai dengan nama module di go.mod Anda
	"go-blockchain-api/internal/api"
	"go-blockchain-api/internal/blockchain"
	"go-blockchain-api/internal/config"
	"go-blockchain-api/internal/engine/aggregator"
	"go-blockchain-api/internal/engine/hasher"
	"go-blockchain-api/internal/models"
	"go-blockchain-api/internal/modules/audit"
	"go-blockchain-api/internal/modules/auth"
	"go-blockchain-api/internal/modules/ingestion"

	"github.com/redis/go-redis/v9"
)

// startPipelineWorker adalah mesin yang berjalan di background setiap 10 detik
func startPipelineWorker(db *gorm.DB, fabricSvc *blockchain.FabricService, redisClient *redis.Client) {
	hashEngine := &hasher.HasherEngine{DB: db}
	aggEngine := &aggregator.AggregatorEngine{DB: db}
	ctx := context.Background()

	go func() {
		log.Println("⚙️ Background Pipeline Worker mulai berjalan...")
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// [Langkah 2 & 3 & 4 tetap sama: Proses log yang sudah ada di DB]
			hashEngine.ProcessPendingLogs()
			aggEngine.ProcessBatch(5)
			if fabricSvc != nil {
				fabricSvc.AnchorPendingRoots()
			}
		}
	}()

	if redisClient == nil {
		return
	}

	go func() {
		log.Println("📥 Redis Queue Worker mulai berjalan...")
		for {
			// BLPop akan menunggu sampai ada item baru, sehingga worker tidak perlu polling tiap 1 detik.
			result, err := redisClient.BLPop(ctx, 0, "audit_log_queue").Result()
			if err != nil {
				log.Printf("⚠️ Error membaca dari Redis: %v\n", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if len(result) < 2 {
				continue
			}

			// Jika ada data, kembalikan dari JSON menjadi Struct
			var logData models.AuditLog
			if err := json.Unmarshal([]byte(result[1]), &logData); err != nil {
				log.Printf("⚠️ Gagal parse log dari Redis: %v\n", err)
				continue
			}

			// Simpan ke PostgreSQL
			if err := db.Create(&logData).Error; err != nil {
				log.Printf("⚠️ Gagal memindah log %s dari Redis ke PostgreSQL: %v\n", logData.HashValue, err)
			} else {
				log.Printf("📥 Memindahkan log %s dari Redis ke Database\n", logData.HashValue)
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

	redisClient := config.ConnectRedis()

	fabricSvc, _ := blockchain.InitFabricGateway(db)

	// 3. Inisialisasi koneksi ke Hyperledger Fabric
	fabricSvc, err := blockchain.InitFabricGateway(db)
	if err != nil {
		log.Println("⚠️ PERINGATAN: Gagal terhubung ke Fabric Gateway!")
		log.Printf("🔍 DETAIL ERROR: %v\n", err)
	}

	// 4. Nyalakan mesin background worker
	startPipelineWorker(db, fabricSvc, redisClient)

	// 5. Inisialisasi Repository & Handler
	auditRepo := audit.NewAuditRepository(db)
	authHandler := &auth.Handler{DB: db}
	ingestionService := &ingestion.Service{Redis: redisClient}
	ingestionHandler := &ingestion.Handler{Service: ingestionService}

	auditHandler := &audit.Handler{
		Repo:   auditRepo,
		Fabric: fabricSvc,
	}

	// 6. Pasang Router yang sudah kita pisahkan ke folder api/
	router := api.SetupRouter(ingestionHandler, auditHandler, authHandler)

	// 7. Jalankan Server API
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🚀 AuditChain Gateway API berjalan di port %s...\n", port)
	router.Run(":" + port)
}
