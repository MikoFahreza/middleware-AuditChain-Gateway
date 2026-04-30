package ingestion

import (
	"fmt"
	"net/http"

	"go-blockchain-api/internal/engine"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	Service *Service
	DB      *gorm.DB
}

// ReceiveLog menerima kumpulan (array) log mentah dari sistem eksternal secara dinamis.
// @Summary Bulk Ingestion Log Audit
// @Description Menerima raw log audit dalam bentuk Array dan memasukkannya ke antrean Redis secara asinkron.
// @Tags Ingestion
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body array true "Array Payload Raw Log Dinamis dari Klien"
// @Success 202 {object} map[string]interface{} "Log diterima"
// @Router /v1/logs [post]
func (h *Handler) ReceiveLog(c *gin.Context) {
	// 1. Ambil Client ID dari hasil kerja Middleware APIKeyAuth
	clientIDVal, exists := c.Get("client_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Identitas klien tidak ditemukan oleh sistem"})
		return
	}
	clientID := clientIDVal.(string)

	// 2. BIND ARRAY: Ubah menjadi slice (array) dari map
	var dynamicPayloads []map[string]interface{}
	if err := c.ShouldBindJSON(&dynamicPayloads); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid, harus berupa Array Objek (Bulk)"})
		return
	}

	if len(dynamicPayloads) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Array log kosong"})
		return
	}

	// 3. OPTIMASI DB: Ambil konfigurasi mapping klien HANYA 1 KALI untuk seluruh batch
	var mapping engine.ClientFieldMapping
	err := h.DB.Table("clients").
		Select("actor_field, action_field, resource_field, data_hash_field, source_system_field").
		Where("id = ?", clientID).
		Scan(&mapping).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil konfigurasi pemetaan klien"})
		return
	}

	// 4. Proses Log melalui Service Layer menggunakan Looping
	var successCount int
	var errorCount int

	for _, payload := range dynamicPayloads {
		// Transformasi per item
		input, err := engine.MapDynamicPayload(payload, &mapping)
		if err != nil {
			// 👇 TAMBAHKAN LOG PRINT INI
			fmt.Printf("❌ [ERROR MAPPING]: %v | Payload: %+v\n", err, payload)
			errorCount++
			continue
		}

		// Sisipkan Client ID secara paksa
		input.ClientID = clientID

		// Masukkan ke Service
		_, err = h.Service.ProcessLog(input)
		if err != nil {
			// 👇 TAMBAHKAN LOG PRINT INI
			fmt.Printf("❌ [ERROR SERVICE]: %v\n", err)
			errorCount++
			continue
		}

		successCount++
	}

	// 5. Kembalikan respons rangkuman proses Bulk
	c.JSON(http.StatusAccepted, gin.H{
		"message":        "Proses bulk ingestion selesai",
		"total_received": len(dynamicPayloads),
		"total_success":  successCount,
		"total_failed":   errorCount,
	})
}
