// @title AuditChain Gateway API
// @version 1.0
// @description API Enterprise untuk sistem audit log berbasis Blockchain dan Merkle Tree.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@auditchain.local

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan token dengan format: Bearer {token}

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name api-key

// @host localhost:8080
// @BasePath /api
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"go-blockchain-api/internal/api"
	"go-blockchain-api/internal/blockchain"
	"go-blockchain-api/internal/config"
	"go-blockchain-api/internal/engine"
	"go-blockchain-api/internal/models"
	"go-blockchain-api/internal/modules/audit"
	"go-blockchain-api/internal/modules/auth"
	"go-blockchain-api/internal/modules/client"
	"go-blockchain-api/internal/modules/ingestion"

	"github.com/redis/go-redis/v9"
)

// startPipelineWorker adalah mesin yang berjalan di background
func startPipelineWorker(db *gorm.DB, fabricSvc *blockchain.FabricService, redisClient *redis.Client) {
	hashEngine := &engine.HasherEngine{DB: db}
	aggEngine := &engine.AggregatorEngine{DB: db}
	ctx := context.Background()

	go func() {
		log.Println("⚙️ Background Pipeline Worker mulai berjalan...")
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			hashEngine.ProcessPendingLogs()
			aggEngine.ProcessBatch(10)
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
			result, err := redisClient.BLPop(ctx, 0, "audit_log_queue").Result()
			if err != nil {
				log.Printf("⚠️ Error membaca dari Redis: %v\n", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if len(result) < 2 {
				continue
			}

			var logData models.AuditLog
			if err := json.Unmarshal([]byte(result[1]), &logData); err != nil {
				log.Printf("⚠️ Gagal parse log dari Redis: %v\n", err)
				continue
			}

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

	// 2. Koneksi ke Infrastruktur
	db := config.ConnectDB()
	redisClient := config.ConnectRedis()

	// 3. Tes Inisialisasi koneksi ke Hyperledger Fabric
	fabricSvc, err := blockchain.InitFabricGateway(db)
	if err != nil {
		log.Println("⚠️ PERINGATAN: Gagal terhubung ke Fabric Gateway!")
		log.Printf("🔍 DETAIL ERROR: %v\n", err)
	}

	// 4. Mulai Background Worker
	startPipelineWorker(db, fabricSvc, redisClient)

	// =========================================================================
	// 5. INISIALISASI MODULE CLEAN ARCHITECTURE
	// =========================================================================

	// A. Modul Audit
	auditRepo := audit.NewAuditRepository(db)
	auditService := audit.NewService(auditRepo, fabricSvc)
	auditHandler := audit.NewHandler(auditService)

	// B. Modul Auth
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := &auth.Handler{Service: authService}

	// C. Modul Ingestion (Antrean)
	ingestionRepo := ingestion.NewRepository(redisClient)
	ingestionService := ingestion.NewService(ingestionRepo)
	ingestionHandler := &ingestion.Handler{
		Service: ingestionService,
		DB:      db,
	}

	// D. Modul Client
	clientRepo := client.NewRepository(db)
	clientService := client.NewService(clientRepo)
	clientHandler := client.NewHandler(clientService)

	// =========================================================================

	// 6. Pasang Router
	router := api.SetupRouter(ingestionHandler, auditHandler, authHandler, clientHandler)

	// 7. Jalankan Server API
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 AuditChain Gateway API berjalan di port %s...\n", port)

	// Gunakan 0.0.0.0 agar API bisa ditembak dari luar Docker (Postman lokal)
	router.Run("0.0.0.0:" + port)
}
