package ingestion

import (
	"context"
	"encoding/json"
	"fmt"

	"go-blockchain-api/internal/engine"
	"go-blockchain-api/internal/models"
)

type Service struct {
	Repo QueueRepository
}

func NewService(repo QueueRepository) *Service {
	return &Service{Repo: repo}
}

// ProcessLog melakukan normalisasi data dan memasukkannya ke antrean Redis secara instan
func (s *Service) ProcessLog(input engine.RawLogInput) (*models.AuditLog, error) {
	// 1. Normalisasi Log (Secara otomatis memberikan status "RECEIVED")
	standardLog, err := engine.Normalize(input)
	if err != nil {
		return nil, fmt.Errorf("gagal menormalisasi log: %v", err)
	}

	// Injeksi ClientID dari JWT/API Key
	standardLog.ClientID = input.ClientID

	logJSON, _ := json.Marshal(standardLog)
	ctx := context.Background()

	err = s.Repo.PushToQueue(ctx, "audit_log_queue", logJSON)
	if err != nil {
		return nil, fmt.Errorf("gagal memasukkan log ke antrean Redis: %v", err)
	}

	return standardLog, nil
}
