package config

import (
	"go-blockchain-api/internal/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal koneksi ke database: %v", err)
	}

	// Auto-Migrate: Membuat tabel jika belum ada sesuai skema model [cite: 236-244]
	err = db.AutoMigrate(
		&models.AuditLog{},
		&models.MerkleMetadata{},
		&models.MerkleProof{},
	)
	if err != nil {
		log.Fatalf("Gagal melakukan migrasi database: %v", err)
	}

	log.Println("✅ Database terhubung dan Off-chain Indexing Schema telah di-migrate.")
	return db
}
