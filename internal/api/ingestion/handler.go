package ingestion

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-blockchain-api/internal/engine/normalizer"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	Redis *redis.Client
}

func (h *Handler) ReceiveLog(c *gin.Context) {
	var input normalizer.RawLogInput

	// 1. Terima dan binding JSON dari request body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	// 2. Normalisasi
	standardLog, err := normalizer.Normalize(input)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Gagal menormalisasi log"})
		return
	}

	// 3. Injeksi Anti-Duplicate Hash (seperti sebelumnya)
	uniqueID := uuid.New().String()
	timestampNano := time.Now().UnixNano()
	uniqueRawString := fmt.Sprintf("%s-%s-%d", standardLog.HashValue, uniqueID, timestampNano)

	hashBytes := sha256.Sum256([]byte(uniqueRawString))
	standardLog.HashValue = hex.EncodeToString(hashBytes[:])

	// ----------------------------------------------------------------------
	// [BARU] SIMPAN KE REDIS QUEUE, BUKAN KE POSTGRESQL
	// ----------------------------------------------------------------------
	// Ubah struct log menjadi JSON string agar bisa disimpan di Redis
	logJSON, _ := json.Marshal(standardLog)

	// Masukkan ke antrean paling belakang (Right Push / RPUSH)
	ctx := context.Background()
	err = h.Redis.RPush(ctx, "audit_log_queue", logJSON).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memasukkan log ke antrean Redis"})
		return
	}
	// ----------------------------------------------------------------------

	// 4. Berikan respons sukses (Jauh lebih cepat dari sebelumnya!)
	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Log berhasil masuk antrean Redis dan akan segera diproses",
		"log_id":     standardLog.LogID,
		"hash_value": standardLog.HashValue,
	})
}
