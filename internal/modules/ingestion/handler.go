package ingestion

import (
	"net/http"

	"go-blockchain-api/internal/engine"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	Service *Service
	DB      *gorm.DB
}

// ReceiveLog menerima log mentah dari sistem eksternal secara dinamis.
// @Summary Ingestion Log Audit
// @Description Menerima raw log audit (dengan struktur dinamis) dan memasukkannya ke antrean Redis secara asinkron.
// @Tags Ingestion
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body object true "Payload Raw Log Dinamis dari Klien"
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

	// 2. Bind JSON Payload dari Klien secara dinamis menggunakan map[string]interface{}
	var dynamicPayload map[string]interface{}
	if err := c.ShouldBindJSON(&dynamicPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	// 3. Ambil konfigurasi mapping klien dari database
	var mapping engine.ClientFieldMapping
	// Menggunakan GORM untuk mengambil mapping dari tabel klien.
	err := h.DB.Table("clients").
		Select("actor_field, action_field, resource_field, data_hash_field").
		Where("id = ?", clientID).
		Scan(&mapping).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil konfigurasi pemetaan klien"})
		return
	}

	// 4. Transformasi Dynamic JSON menjadi RawLogInput yang baku
	input, err := engine.MapDynamicPayload(dynamicPayload, &mapping)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal memetakan struktur data log"})
		return
	}

	// 5. Sisipkan Client ID ke dalam Payload secara paksa
	// (Klien tidak bisa memalsukan ID mereka dari JSON body)
	input.ClientID = clientID

	// 6. Proses Log melalui Service Layer
	// ProcessLog sekarang menerima 'input' yang sudah dinormalisasi dan di-mapping
	standardLog, err := h.Service.ProcessLog(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Log berhasil masuk antrean Redis dan akan segera diproses",
		"log_id":  standardLog.LogID,
	})
}
