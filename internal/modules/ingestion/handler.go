package ingestion

import (
	"net/http"

	"go-blockchain-api/internal/engine/normalizer"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

// ReceiveLog menerima log mentah dari sistem eksternal.
// @Summary Ingestion Log Audit
// @Description Menerima raw log audit dan memasukkannya ke antrean Redis secara asinkron.
// @Tags Ingestion
// @Accept json
// @Produce json
// @Security ApiKeyAuth  <-- 👇 TAMBAHKAN BARIS INI
// @Param request body normalizer.RawLogInput true "Payload Raw Log"
// @Success 202 {object} map[string]interface{} "Log diterima"
// @Router /v1/logs [post]
func (h *Handler) ReceiveLog(c *gin.Context) {
	var input normalizer.RawLogInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

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
