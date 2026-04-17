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

// ReceiveLog menerima log mentah dari sistem eksternal.
// @Summary Ingestion Log Audit
// @Description Menerima raw log audit dan memasukkannya ke antrean Redis secara asinkron.
// @Tags Ingestion
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body engine.RawLogInput true "Payload Raw Log"
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

	var input engine.RawLogInput

	// 2. Bind JSON Payload dari Klien
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	// 3. Sisipkan Client ID ke dalam Payload secara paksa
	// (Klien tidak bisa memalsukan ID mereka dari JSON body)
	input.ClientID = clientID

	// 4. Proses Log melalui Service Layer
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
