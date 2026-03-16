package audit

import (
	"go-blockchain-api/pkg/crypto"
	"net/http"

	"encoding/json"
	"go-blockchain-api/internal/blockchain"

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
	txHash := c.Param("hash")

	// 1. Ambil data log berdasarkan Hash
	logData, err := h.Repo.GetLogByHash(txHash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data log transaksi tidak ditemukan"})
		return
	}

	// 2. Jika belum masuk ke Merkle Tree, tidak bisa diverifikasi
	if logData.MerkleRoot == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "PENDING",
			"message": "Log ini belum diagregasi ke dalam Merkle Tree",
		})
		return
	}

	// 3. Ambil jalur Merkle Proof (Sibling Hashes) dari database [cite: 201, 243-244]
	proofRecords, _ := h.Repo.GetProofsByHash(txHash)
	var siblingHashes []string
	for _, p := range proofRecords {
		siblingHashes = append(siblingHashes, p.SiblingHash)
	}

	// 4. Lakukan verifikasi matematis secara off-chain [cite: 229-230]
	isValid := crypto.VerifyMerkleProof(txHash, siblingHashes, logData.MerkleRoot)

	// Respons Audit
	c.JSON(http.StatusOK, gin.H{
		"transaction_hash": txHash,
		"merkle_root":      logData.MerkleRoot,
		"blockchain_tx_id": logData.BlockchainTxID,
		"is_valid":         isValid,
		"proof_path":       siblingHashes,
		"message":          "Verifikasi selesai",
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
