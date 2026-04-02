package normalizer

import (
	"encoding/json"
	"errors"
	"go-blockchain-api/internal/models"
	"time"

	"github.com/google/uuid"
)

// RawLogInput adalah representasi data mentah yang dikirim oleh sistem klien
type RawLogInput struct {
	LogID                string `json:"log_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Actor                string `json:"actor" binding:"required" example:"auditor_utama"`
	Action               string `json:"action" binding:"required" example:"UPDATE_SALARY"`
	Resource             string `json:"resource" binding:"required" example:"Table_Employees"`
	Timestamp            string `json:"timestamp" example:"2024-06-01T15:04:05Z07:00"`
	SourceSystem         string `json:"source_system" example:"HRIS_App_v2"`
	AuthorizationContext string `json:"authorization_context" example:"Role: Admin"`
	// Contoh metadata dinamis
	Metadata map[string]interface{} `json:"metadata" example:"{\"ip_address\":\"192.168.1.45\", \"browser\":\"Chrome\"}"`
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
		Status:               "RECEIVED",
	}

	return standardLog, nil
}
