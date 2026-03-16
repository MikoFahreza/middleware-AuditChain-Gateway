package ingestion

import (
	"net/http"

	"go-blockchain-api/internal/engine/normalizer"

	"github.com/gin-gonic/gin"
)

// Handler adalah "Pelayan" yang menangani request/response HTTP
type Handler struct {
	Service *Service // Pelayan harus tahu siapa Kokinya
}

func (h *Handler) ReceiveLog(c *gin.Context) {
	var input normalizer.RawLogInput

	// 1. Terima JSON dari Postman/Klien
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	// 2. Serahkan bahan mentah ke Dapur (Service)
	standardLog, err := h.Service.ProcessLog(input)
	if err != nil {
		// Jika koki gagal memasak, beritahu klien
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Berikan respons sukses ke klien
	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Log berhasil masuk antrean Redis dan akan segera diproses",
		"log_id":     standardLog.LogID,
		"hash_value": standardLog.HashValue,
	})
}
