package client

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"go-blockchain-api/internal/models"

	"github.com/google/uuid"
)

type Service interface {
	// Signature diubah untuk menerima struct CreateClientRequest
	RegisterClient(req CreateClientRequest) (*models.Client, string, error)
}

type clientService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &clientService{repo: repo}
}

func (s *clientService) RegisterClient(req CreateClientRequest) (*models.Client, string, error) {
	// 1. Generate Raw API Key
	randomUUID := uuid.New().String()
	rawAPIKey := "ak_live_" + hex.EncodeToString([]byte(randomUUID))[:16]

	// 2. Hash API Key untuk keamanan Database
	hash := sha256.Sum256([]byte(rawAPIKey))
	apiKeyHash := hex.EncodeToString(hash[:])

	// 3. Konfigurasi Nilai Default (Jika payload dari Klien kosong)
	tier := req.SubscriptionTier
	if tier == "" {
		tier = "basic"
	}

	rateLimit := req.RateLimitPerSec
	if rateLimit == 0 {
		rateLimit = 50
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	// 4. Buat objek Client baru beserta field pemetaannya
	newClient := &models.Client{
		CompanyName:      req.CompanyName,
		APIKeyHash:       apiKeyHash,
		APIKeyPrefix:     rawAPIKey[:10],
		SubscriptionTier: tier,
		RateLimitPerSec:  rateLimit,
		Status:           status,
		ActorField:       req.ActorField,
		ActionField:      req.ActionField,
		ResourceField:    req.ResourceField,
	}

	// 5. Simpan ke Database
	if err := s.repo.CreateClient(newClient); err != nil {
		return nil, "", errors.New("gagal mendaftarkan klien ke database")
	}

	// Kembalikan data klien dan Raw API Key (untuk ditampilkan HANYA SEKALI ke user)
	return newClient, rawAPIKey, nil
}
