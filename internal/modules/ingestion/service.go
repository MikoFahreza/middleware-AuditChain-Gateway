package ingestion

import (
	"context"

	"encoding/json"
	"fmt"
	"time"

	"go-blockchain-api/internal/engine"
	"go-blockchain-api/internal/models"
	"go-blockchain-api/pkg/crypto"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	Redis *redis.Client
}

// ProcessLog melakukan normalisasi, hashing, dan memasukkan ke antrean Redis
func (s *Service) ProcessLog(input engine.RawLogInput) (*models.AuditLog, error) {
	// 1. Normalisasi Log
	standardLog, err := engine.Normalize(input)
	if err != nil {
		return nil, fmt.Errorf("gagal menormalisasi log: %v", err)
	}

	// 2. Injeksi Anti-Duplicate Hash
	uniqueID := uuid.New().String()
	timestampNano := time.Now().UnixNano()
	uniqueRawString := fmt.Sprintf("%s-%s-%d", standardLog.HashValue, uniqueID, timestampNano)

	standardLog.HashValue = crypto.GenerateSHA3_256(uniqueRawString)

	// 3. Simpan ke Redis Queue
	logJSON, _ := json.Marshal(standardLog)
	ctx := context.Background()

	err = s.Redis.RPush(ctx, "audit_log_queue", logJSON).Err()
	if err != nil {
		return nil, fmt.Errorf("gagal memasukkan log ke antrean Redis: %v", err)
	}

	return standardLog, nil
}
