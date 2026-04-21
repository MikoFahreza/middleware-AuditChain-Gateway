package client

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"go-blockchain-api/internal/models"

	"github.com/google/uuid"
)

type Service interface {
	RegisterClient(companyName string) (*models.Client, string, error)
}

type clientService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &clientService{repo: repo}
}

func (s *clientService) RegisterClient(companyName string) (*models.Client, string, error) {
	// 1. Generate Raw API Key
	randomUUID := uuid.New().String()
	rawAPIKey := "ak_live_" + hex.EncodeToString([]byte(randomUUID))[:16]

	// 2. Hash API Key untuk keamanan Database
	hash := sha256.Sum256([]byte(rawAPIKey))
	apiKeyHash := hex.EncodeToString(hash[:])

	// 3. Buat objek Client baru
	newClient := &models.Client{
		CompanyName:  companyName,
		APIKeyHash:   apiKeyHash,
		APIKeyPrefix: rawAPIKey[:10],
		Status:       "active",
	}

	// 4. Simpan ke Database
	if err := s.repo.CreateClient(newClient); err != nil {
		return nil, "", errors.New("gagal mendaftarkan klien ke database")
	}

	// Kembalikan data klien dan Raw API Key (untuk ditampilkan HANYA SEKALI ke user)
	return newClient, rawAPIKey, nil
}
