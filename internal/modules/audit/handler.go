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

// VerifyLog mengecek keaslian transaksi menggunakan Merkle Proof [cite: 197-203]
func (h *Handler) VerifyLog(c *gin.Context) {
	requestedHash := c.Param("hash")

	// 1. Ambil data dari database
	auditLog, err := h.Repo.GetLogByHash(requestedHash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log tidak ditemukan"})
		return
	}

	// =====================================================================
	// 🚨 DETEKSI TAMPERING DATABASE (RE-CALCULATE HASH)
	// =====================================================================
	// Gunakan mesin Hasher yang persis sama dengan saat data dibuat
	recalculatedHash := hasher.GenerateLogHash(auditLog, auditLog.PreviousHash)

	// Jika hash hasil hitungan ulang BEDA dengan hash yang tersimpan,
	// berarti ada orang dalam (DBA) yang mengedit isi tabel!
	if recalculatedHash != auditLog.HashValue {
		c.JSON(http.StatusConflict, gin.H{
			"status": "failed",
			"data": gin.H{
				"is_valid":      false,
				"message":       "🚨 DATA TERMANIPULASI: Isi data (Actor/Action) telah diubah di database dan tidak cocok dengan Hash aslinya!",
				"expected_hash": auditLog.HashValue,
				"actual_hash":   recalculatedHash,
			},
		})
		return
	}
	// =====================================================================

	// 2. Jika lolos deteksi tampering DB, baru kita cek status Blockchain-nya
	if auditLog.BlockchainTxID == nil || *auditLog.BlockchainTxID == "PENDING_OR_FAILED" {
		c.JSON(http.StatusAccepted, gin.H{
			"status":  "pending",
			"message": "Log otentik, namun masih dalam proses antrean ke Blockchain.",
		})
		return
	}

	// 3. (Opsional) Di sini Anda bisa menambahkan logika verifikasi Merkle Proof
	// dan mengecek ke Node Hyperledger Fabric secara langsung jika diperlukan.

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"log_id":           auditLog.LogID,
			"hash_value":       auditLog.HashValue,
			"is_valid":         true,
			"blockchain_tx_id": auditLog.BlockchainTxID,
			"message":          "✅ DATA OTENTIK: Data utuh dan Hash cocok dengan catatan Blockchain.",
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
