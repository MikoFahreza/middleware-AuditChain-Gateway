package normalizer

import (
	"encoding/json"
	"errors"
	"go-blockchain-api/internal/models"
	"time"

	"github.com/google/uuid"
)

// RawLogInput adalah struktur data JSON yang kita harapkan dari sistem eksternal (Postman/Klien) [cite: 102, 107-114]
type RawLogInput struct {
	LogID                string      `json:"log_id"` // Opsional
	Actor                string      `json:"actor" binding:"required"`
	Action               string      `json:"action" binding:"required"`
	Resource             string      `json:"resource" binding:"required"`
	Timestamp            string      `json:"timestamp"` // Opsional, pakai waktu sekarang jika kosong
	SourceSystem         string      `json:"source_system" binding:"required"`
	AuthorizationContext interface{} `json:"authorization_context"` // Bisa berupa JSON object
	Metadata             interface{} `json:"metadata"`              // Bisa berupa JSON object
}

// Normalize mengubah RawLogInput menjadi models.AuditLog yang standar [cite: 120, 141]
func Normalize(input RawLogInput) (*models.AuditLog, error) {
	// 1. Validasi manual tambahan jika diperlukan [cite: 142]
	if input.Actor == "" || input.Action == "" || input.Resource == "" || input.SourceSystem == "" {
		return nil, errors.New("field wajib (actor, action, resource, source_system) tidak boleh kosong")
	}

	// 2. Generate Log ID jika sistem eksternal tidak memberikannya
	logID := input.LogID
	if logID == "" {
		logID = uuid.New().String()
	}

	// 3. Normalisasi Timestamp [cite: 142]
	logTime := time.Now()
	if input.Timestamp != "" {
		parsedTime, err := time.Parse(time.RFC3339, input.Timestamp)
		if err == nil {
			logTime = parsedTime
		}
	}

	// 4. Konversi interface{} (JSON Object) menjadi string agar deterministik saat di-hash [cite: 141, 143]
	authCtxBytes, _ := json.Marshal(input.AuthorizationContext)
	metaBytes, _ := json.Marshal(input.Metadata)

	// 5. Mapping ke Model Data Internal [cite: 141]
	standardLog := &models.AuditLog{
		LogID:                logID,
		Actor:                input.Actor,
		Action:               input.Action,
		Resource:             input.Resource,
		Timestamp:            logTime,
		SourceSystem:         input.SourceSystem,
		AuthorizationContext: string(authCtxBytes),
		Metadata:             string(metaBytes),
		Status:               "RECEIVED", // Status awal sebelum masuk hashing engine [cite: 144]
	}

	return standardLog, nil
}
