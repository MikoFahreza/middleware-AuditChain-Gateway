package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-blockchain-api/internal/models"
	"time"

	"github.com/google/uuid"
)

// ClientFieldMapping adalah kamus pemetaan field khusus untuk klien tertentu.
// Ini biasanya diambil dari database berdasarkan ClientID yang sedang login.
type ClientFieldMapping struct {
	ActorField    string `json:"actor_field"`    // misal: "user_name"
	ActionField   string `json:"action_field"`   // misal: "event_type"
	ResourceField string `json:"resource_field"` // misal: "table_name"
}

// RawLogInput adalah representasi data mentah yang dikirim oleh sistem klien
type RawLogInput struct {
	ClientID             string                 `json:"-"`
	LogID                string                 `json:"log_id"`
	Actor                string                 `json:"actor"`
	Action               string                 `json:"action"`
	Resource             string                 `json:"resource"`
	Timestamp            string                 `json:"timestamp"`
	SourceSystem         string                 `json:"source_system"`
	AuthorizationContext map[string]interface{} `json:"authorization_context"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// 👇 FUNGSI BARU: MapDynamicPayload
// Fungsi ini menerjemahkan JSON dinamis dari klien menjadi RawLogInput yang baku
func MapDynamicPayload(dynamicPayload map[string]interface{}, mapping *ClientFieldMapping) (RawLogInput, error) {
	var input RawLogInput

	getString := func(key string) string {
		if val, ok := dynamicPayload[key]; ok {
			return fmt.Sprintf("%v", val)
		}
		return ""
	}

	// Kunci default
	keyActor := "actor"
	keyAction := "action"
	keyResource := "resource"

	// Override jika mapping ada
	if mapping != nil {
		if mapping.ActorField != "" {
			keyActor = mapping.ActorField
		}
		if mapping.ActionField != "" {
			keyAction = mapping.ActionField
		}
		if mapping.ResourceField != "" {
			keyResource = mapping.ResourceField
		}
	}

	input.Actor = getString(keyActor)
	input.Action = getString(keyAction)
	input.Resource = getString(keyResource)

	// Field lainnya
	input.LogID = getString("log_id")
	input.Timestamp = getString("timestamp")
	input.SourceSystem = getString("source_system")

	// 👇 INI BAGIAN PALING KRUSIAL UNTUK MENANGKAP METADATA 👇
	if meta, exists := dynamicPayload["metadata"]; exists {
		if metaMap, ok := meta.(map[string]interface{}); ok {
			input.Metadata = metaMap
		}
	}

	return input, nil
}

// Normalize mengubah RawLogInput menjadi models.AuditLog yang standar (SAMA SEPERTI MILIK ANDA)
func Normalize(input RawLogInput) (*models.AuditLog, error) {
	// 1. Validasi manual tambahan jika diperlukan
	if input.Actor == "" || input.Action == "" || input.Resource == "" || input.SourceSystem == "" {
		return nil, errors.New("field wajib (actor, action, resource, source_system) tidak boleh kosong")
	}

	// 2. Generate Log ID jika sistem eksternal tidak memberikannya
	logID := input.LogID
	if logID == "" {
		logID = uuid.New().String()
	}

	// 3. Normalisasi Timestamp
	logTime := time.Now()
	if input.Timestamp != "" {
		parsedTime, err := time.Parse(time.RFC3339, input.Timestamp)
		if err == nil {
			logTime = parsedTime
		}
	}

	// 4. Konversi interface{} (JSON Object) menjadi string agar deterministik saat di-hash
	authCtxBytes, _ := json.Marshal(input.AuthorizationContext)
	metaBytes, _ := json.Marshal(input.Metadata)

	// 5. Mapping ke Model Data Internal
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
