package audit

import (
	"encoding/json"
	"errors"

	"go-blockchain-api/internal/blockchain"
	"go-blockchain-api/internal/engine"
	"go-blockchain-api/internal/models"
	"go-blockchain-api/pkg/crypto"
)

// VerificationResult adalah struktur laporan hasil pengecekan integritas log
type VerificationResult struct {
	Status       string
	Message      string
	IsValid      bool
	ExpectedHash string
	ActualHash   string
	DBRoot       string
	ChainRoot    string
	LogID        string
	TxID         *string
}

// DataVerificationResult adalah struktur laporan hasil pengecekan integritas data klien
type DataVerificationResult struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	IsValid      bool   `json:"is_valid"`
	Resource     string `json:"resource"`
	ExpectedHash string `json:"expected_data_hash"`
	ActualHash   string `json:"actual_data_hash"`
	LastLogID    string `json:"last_log_id"`
}

type Service interface {
	GetDashboardStats() (map[string]int64, error)
	VerifyLogIntegrity(hash string) (*VerificationResult, error)
	GetFabricRecord(anchorID string) (map[string]interface{}, error)
	VerifyDataIntegrity(resource string, rawData *map[string]interface{}) (*DataVerificationResult, error)
	GetRecentLogs(limit int) ([]models.AuditLog, error)
	GetResourceInventory() ([]models.AuditLog, error)
}

type auditService struct {
	repo   AuditRepository
	fabric *blockchain.FabricService
}

func NewService(repo AuditRepository, fabric *blockchain.FabricService) Service {
	return &auditService{
		repo:   repo,
		fabric: fabric,
	}
}

func (s *auditService) GetDashboardStats() (map[string]int64, error) {
	return s.repo.GetDashboardStats()
}

func (s *auditService) VerifyLogIntegrity(hash string) (*VerificationResult, error) {
	// 1. Ambil data dari Database
	auditLog, err := s.repo.GetLogByHash(hash)
	if err != nil {
		return nil, errors.New("log_not_found")
	}

	// 2. Cek Manipulasi Lokal (Re-Hashing)
	recalculatedHash := engine.GenerateLogHash(auditLog, auditLog.PreviousHash)
	if recalculatedHash != auditLog.HashValue {
		return &VerificationResult{
			Status:       "failed_local",
			Message:      "🚨 DATA TERMANIPULASI: Isi data telah diubah di database dan tidak cocok dengan Hash aslinya!",
			IsValid:      false,
			ExpectedHash: auditLog.HashValue,
			ActualHash:   recalculatedHash,
		}, nil
	}

	// 3. Cek Status Antrean Blockchain
	if auditLog.BlockchainTxID == nil || *auditLog.BlockchainTxID == "PENDING_OR_FAILED" {
		return &VerificationResult{
			Status:  "pending",
			Message: "Log otentik secara lokal, namun masih dalam proses antrean ke Blockchain.",
			IsValid: true,
		}, nil
	}

	// 4. Tarik dari Ledger Hyperledger Fabric
	onChainData, err := s.fabric.GetAnchorFromLedger(*auditLog.BlockchainTxID)
	if err != nil {
		return nil, errors.New("fabric_error")
	}

	var fabricResponse struct {
		MerkleRoot string `json:"merkle_root"`
	}
	if err := json.Unmarshal([]byte(onChainData), &fabricResponse); err != nil {
		return nil, errors.New("parse_error")
	}

	// 5. Verifikasi Merkle Root DB vs On-Chain
	if fabricResponse.MerkleRoot != auditLog.MerkleRoot {
		return &VerificationResult{
			Status:    "failed_onchain",
			Message:   "🚨 FATAL MISMATCH: Merkle Root di database TIDAK DIAKUI oleh jaringan Blockchain!",
			IsValid:   false,
			DBRoot:    auditLog.MerkleRoot,
			ChainRoot: fabricResponse.MerkleRoot,
		}, nil
	}

	// 6. Jika semua lolos, data terbukti 100% Otentik
	return &VerificationResult{
		Status:       "success",
		Message:      "✅ DATA OTENTIK 100%: Data utuh dan Merkle Root terverifikasi dengan Ledger Blockchain.",
		IsValid:      true,
		LogID:        auditLog.LogID,
		ExpectedHash: auditLog.HashValue,
		DBRoot:       auditLog.MerkleRoot,
		TxID:         auditLog.BlockchainTxID,
	}, nil
}

func (s *auditService) GetFabricRecord(anchorID string) (map[string]interface{}, error) {
	if s.fabric == nil {
		return nil, errors.New("fabric_bypass")
	}

	fabricDataString, err := s.fabric.GetAnchorFromLedger(anchorID)
	if err != nil {
		return nil, errors.New("fabric_not_found")
	}

	var jsonResponse map[string]interface{}
	if err := json.Unmarshal([]byte(fabricDataString), &jsonResponse); err != nil {
		return nil, errors.New("parse_error")
	}

	return jsonResponse, nil
}

// Logika Verifikasi Data Klien berdasarkan Resource dan Data Aktual yang diberikan
func (s *auditService) VerifyDataIntegrity(resource string, rawData *map[string]interface{}) (*DataVerificationResult, error) {
	// 1. Dapatkan log terakhir
	lastLog, err := s.repo.GetLatestLogByResource(resource)
	if err != nil {
		return nil, errors.New("log_not_found")
	}

	// 2. Identifikasi Kondisi Data Aktual
	var actualHash string
	isDataEmpty := rawData == nil || len(*rawData) == 0

	if !isDataEmpty {
		dataBytes, _ := json.Marshal(*rawData)
		actualHash = crypto.GenerateSHA3_256(string(dataBytes))
	}

	// 3. Evaluasi Berdasarkan Jenis Aksi (Action) Terakhir
	isValid := false
	status := "failed"
	var msg string

	// Cek apakah aksi terakhir mengandung kata DELETE (bisa DELETE, SOFT_DELETE, HARD_DELETE)
	isLastActionDelete := lastLog.Action == "DELETE"

	if isLastActionDelete {
		// LOGIKA UNTUK DELETE
		if isDataEmpty {
			isValid = true
			status = "success"
			msg = "✅ DATA VALID: Log terakhir adalah DELETE, dan data aktual memang sudah tidak ada di database."
		} else {
			msg = "🔴 DATA TERMANIPULASI (GHOST DATA): Log terakhir menyatakan data telah di-DELETE, tetapi data aktual masih ditemukan di database lokal!"
		}
	} else {
		// LOGIKA UNTUK INSERT / UPDATE
		if isDataEmpty {
			msg = "🔴 DATA TERMANIPULASI (ILLEGAL DELETION): Log terakhir tidak mencatat adanya penghapusan, tetapi data aktual tiba-tiba HILANG dari database lokal!"
		} else {
			if actualHash == lastLog.DataHash {
				isValid = true
				status = "success"
				msg = "✅ DATA VALID: Kondisi data aktual sama persis dengan jejak terakhir di sistem audit."
			} else {
				msg = "🔴 DATA TERMANIPULASI (UNAUTHORIZED MODIFICATION): Isi data di database lokal saat ini BERBEDA dengan jejak sah terakhir."
			}
		}
	}

	return &DataVerificationResult{
		Status:       status,
		Message:      msg,
		IsValid:      isValid,
		Resource:     resource,
		ExpectedHash: lastLog.DataHash,
		ActualHash:   actualHash,
		LastLogID:    lastLog.LogID,
	}, nil
}

func (s *auditService) GetRecentLogs(limit int) ([]models.AuditLog, error) {
	return s.repo.GetRecentLogs(limit)
}

func (s *auditService) GetResourceInventory() ([]models.AuditLog, error) {
	return s.repo.GetResourceInventory()
}
