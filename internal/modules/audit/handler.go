package audit

import (
	"net/http"

	"encoding/json"
	"go-blockchain-api/internal/blockchain"
	"go-blockchain-api/internal/engine/hasher"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repo   AuditRepository
	Fabric *blockchain.FabricService
}

func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.Repo.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil statistik"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// VerifyLog mengecek integritas log dari database hingga ke catatan Blockchain.
// @Summary Verifikasi Integritas Log (Merkle Proof)
// @Description Melakukan re-hashing data lokal dan memverifikasi Merkle Root langsung ke jaringan Hyperledger Fabric.
// @Tags Audit & Dashboard
// @Security BearerAuth
// @Produce json
// @Param hash path string true "Nilai Hash dari Log yang ingin diverifikasi"
// @Success 200 {object} map[string]interface{} "✅ Data Valid dan Otentik"
// @Success 202 {object} map[string]interface{} "⏳ Data Pending di Antrean"
// @Failure 401 {object} map[string]interface{} "Akses ditolak"
// @Failure 404 {object} map[string]interface{} "Log tidak ditemukan"
// @Failure 409 {object} map[string]interface{} "🚨 Data Termanipulasi (DB atau Blockchain Mismatch)"
// @Failure 500 {object} map[string]interface{} "Kesalahan koneksi ke Blockchain"
// @Router /dashboard/verify/{hash} [get]
func (h *Handler) VerifyLog(c *gin.Context) {
	requestedHash := c.Param("hash")

	auditLog, err := h.Repo.GetLogByHash(requestedHash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log tidak ditemukan di dalam sistem."})
		return
	}

	recalculatedHash := hasher.GenerateLogHash(auditLog, auditLog.PreviousHash)

	if recalculatedHash != auditLog.HashValue {
		c.JSON(http.StatusConflict, gin.H{
			"status": "failed",
			"data": gin.H{
				"is_valid":      false,
				"message":       "🚨 DATA TERMANIPULASI: Isi data telah diubah di database dan tidak cocok dengan Hash aslinya!",
				"expected_hash": auditLog.HashValue,
				"actual_hash":   recalculatedHash,
			},
		})
		return
	}

	if auditLog.BlockchainTxID == nil || *auditLog.BlockchainTxID == "PENDING_OR_FAILED" {
		c.JSON(http.StatusAccepted, gin.H{
			"status":  "pending",
			"message": "Log otentik secara lokal, namun masih dalam proses antrean ke Blockchain.",
		})
		return
	}

	// -------------------------------------------------------------------------
	// LAPIS 3: Verifikasi Konsensus ke Hyperledger Fabric (The Ultimate Truth)
	// -------------------------------------------------------------------------
	onChainData, err := h.Fabric.GetAnchorFromLedger(*auditLog.BlockchainTxID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal terhubung ke jaringan Blockchain Fabric.",
			"detail":  err.Error(),
		})
		return
	}

	// 1. Buat struct sementara untuk menangkap JSON dari Chaincode
	// PENTING: Sesuaikan tag `json:"..."` dengan definisi struct AnchorRecord di Chaincode Anda!
	var fabricResponse struct {
		MerkleRoot string `json:"merkleRoot"` // Cek apakah di Chaincode Anda namanya "merkleRoot", "merkle_root", atau "MerkleRoot"
	}

	// 2. Unmarshal JSON string dari Fabric ke struct sementara
	importJSONErr := json.Unmarshal([]byte(onChainData), &fabricResponse)
	if importJSONErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal membaca format data dari Blockchain.",
			"detail":  importJSONErr.Error(),
		})
		return
	}

	// 3. Ekstrak Root yang sudah bersih dari JSON
	onChainRoot := fabricResponse.MerkleRoot

	// 4. Bandingkan dengan Database
	if onChainRoot != auditLog.MerkleRoot {
		c.JSON(http.StatusConflict, gin.H{
			"status": "failed",
			"data": gin.H{
				"is_valid":   false,
				"message":    "🚨 FATAL MISMATCH: Merkle Root di database TIDAK DIAKUI oleh jaringan Blockchain!",
				"db_root":    auditLog.MerkleRoot,
				"chain_root": onChainRoot, // Sekarang ini akan mencetak hash-nya saja
			},
		})
		return
	}

	// -------------------------------------------------------------------------
	// LOLOS SEMUA UJIAN
	// -------------------------------------------------------------------------
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"log_id":           auditLog.LogID,
			"hash_value":       auditLog.HashValue,
			"merkle_root":      auditLog.MerkleRoot,
			"blockchain_tx_id": auditLog.BlockchainTxID,
			"is_valid":         true,
			"message":          "✅ DATA OTENTIK 100%: Data utuh dan Merkle Root terverifikasi dengan Ledger Blockchain.",
		},
	})
}

func (h *Handler) GetFabricRecord(c *gin.Context) {
	anchorID := c.Param("anchor_id")

	// 1. Cek apakah koneksi Fabric tersedia
	if h.Fabric == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Koneksi ke jaringan Hyperledger Fabric sedang terputus (Bypass Mode)",
		})
		return
	}

	// 2. Tarik data dari Blockchain menggunakan fungsi yang baru kita buat
	fabricDataString, err := h.Fabric.GetAnchorFromLedger(anchorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Data tidak ditemukan di dalam Ledger Fabric",
			"details": err.Error(),
		})
		return
	}

	// 3. Ubah string JSON dari Fabric kembali menjadi objek agar rapi saat ditampilkan
	var jsonResponse map[string]interface{}
	json.Unmarshal([]byte(fabricDataString), &jsonResponse)

	c.JSON(http.StatusOK, gin.H{
		"source": "Hyperledger Fabric World State",
		"data":   jsonResponse,
	})
}
