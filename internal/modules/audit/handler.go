package audit

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.Service.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil statistik"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// VerifyLog mengecek integritas log dari database hingga ke catatan Blockchain.
// (Komentar Swagger biarkan utuh seperti milik Anda)
func (h *Handler) VerifyLog(c *gin.Context) {
	requestedHash := c.Param("hash")

	result, err := h.Service.VerifyLogIntegrity(requestedHash)

	// Tangani error sistem
	if err != nil {
		switch err.Error() {
		case "log_not_found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Log tidak ditemukan di dalam sistem."})
		case "fabric_error":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal terhubung ke jaringan Blockchain Fabric."})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan sistem saat memverifikasi data."})
		}
		return
	}

	// Tangani hasil logika bisnis (Business Logic)
	switch result.Status {
	case "failed_local":
		c.JSON(http.StatusConflict, gin.H{
			"status": "failed",
			"data": gin.H{
				"is_valid":      result.IsValid,
				"message":       result.Message,
				"expected_hash": result.ExpectedHash,
				"actual_hash":   result.ActualHash,
			},
		})
	case "pending":
		c.JSON(http.StatusAccepted, gin.H{
			"status":  "pending",
			"message": result.Message,
		})
	case "failed_onchain":
		c.JSON(http.StatusConflict, gin.H{
			"status": "failed",
			"data": gin.H{
				"is_valid":   result.IsValid,
				"message":    result.Message,
				"db_root":    result.DBRoot,
				"chain_root": result.ChainRoot,
			},
		})
	case "success":
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"log_id":           result.LogID,
				"hash_value":       result.ExpectedHash,
				"merkle_root":      result.DBRoot,
				"blockchain_tx_id": result.TxID,
				"is_valid":         result.IsValid,
				"message":          result.Message,
			},
		})
	}
}

func (h *Handler) GetFabricRecord(c *gin.Context) {
	anchorID := c.Param("anchor_id")

	data, err := h.Service.GetFabricRecord(anchorID)
	if err != nil {
		switch err.Error() {
		case "fabric_bypass":
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Koneksi ke jaringan Hyperledger Fabric sedang terputus (Bypass Mode)"})
		case "fabric_not_found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Data tidak ditemukan di dalam Ledger Fabric"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses data dari Fabric"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"source": "Hyperledger Fabric World State",
		"data":   data,
	})
}
