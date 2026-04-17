package audit

import (
	"encoding/json"
	"errors"

	"go-blockchain-api/internal/blockchain"
	"go-blockchain-api/internal/engine"
)

// VerificationResult adalah struktur laporan hasil pengecekan integritas
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

type Service interface {
	GetDashboardStats() (map[string]int64, error)
	VerifyLogIntegrity(hash string) (*VerificationResult, error)
	GetFabricRecord(anchorID string) (map[string]interface{}, error)
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
