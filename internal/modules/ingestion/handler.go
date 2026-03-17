package ingestion

import (
	"net/http"

	"go-blockchain-api/internal/engine/normalizer"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

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
