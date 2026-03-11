package ingestion

import (
	"go-blockchain-api/internal/engine/normalizer"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

// ReceiveLog adalah endpoint POST /logs untuk menerima log transaksi [cite: 101]
func (h *Handler) ReceiveLog(c *gin.Context) {
	var input normalizer.RawLogInput

	// 1. Terima dan binding JSON dari request body [cite: 102]
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Format JSON tidak valid atau ada field wajib yang hilang",
			"details": err.Error(),
		})
		return
	}

	// 2. Kirim ke Normalization Engine [cite: 115, 141]
	standardLog, err := normalizer.Normalize(input)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "Gagal menormalisasi log",
			"details": err.Error(),
		})
		return
	}

	// 3. Simpan sementara ke Buffer / Database dengan status 'RECEIVED' [cite: 144]
	if err := h.DB.Create(standardLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menyimpan log ke database internal",
		})
		return
	}

	// 4. Berikan respons sukses ke sistem eksternal
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Log berhasil diterima dan sedang diproses dalam pipeline",
		"log_id":  standardLog.LogID,
		"status":  standardLog.Status,
	})
}
