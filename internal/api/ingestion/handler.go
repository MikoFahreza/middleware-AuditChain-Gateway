package ingestion

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"go-blockchain-api/internal/engine/normalizer"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

// ReceiveLog adalah endpoint POST /logs untuk menerima log transaksi
func (h *Handler) ReceiveLog(c *gin.Context) {
	var input normalizer.RawLogInput

	// 1. Terima dan binding JSON dari request body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Format JSON tidak valid atau ada field wajib yang hilang",
			"details": err.Error(),
		})
		return
	}

	// 2. Kirim ke Normalization Engine
	standardLog, err := normalizer.Normalize(input)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "Gagal menormalisasi log",
			"details": err.Error(),
		})
		return
	}

	// ----------------------------------------------------------------------
	// [BARU] INJEKSI ANTI-DUPLICATE (DETERMINISM FIX)
	// ----------------------------------------------------------------------
	// Kita menimpa (override) hash dari normalizer dengan hash yang 100% unik
	uniqueID := uuid.New().String()
	timestampNano := time.Now().UnixNano()

	// Gabungkan hasil dari normalizer dengan UUID dan Timestamp
	uniqueRawString := fmt.Sprintf("%s-%s-%d", standardLog.HashValue, uniqueID, timestampNano)

	// Hitung ulang Hash SHA-256
	hashBytes := sha256.Sum256([]byte(uniqueRawString))
	standardLog.HashValue = hex.EncodeToString(hashBytes[:])
	// ----------------------------------------------------------------------

	// 3. Simpan sementara ke Buffer / Database dengan status 'RECEIVED'
	if err := h.DB.Create(standardLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menyimpan log ke database internal",
			"details": err.Error(),
		})
		return
	}

	// 4. Berikan respons sukses ke sistem eksternal
	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Log berhasil diterima dan sedang diproses dalam pipeline",
		"log_id":     standardLog.LogID,
		"status":     standardLog.Status,
		"hash_value": standardLog.HashValue,
	})
}
