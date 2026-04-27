package audit

import (
	"go-blockchain-api/internal/models"

	"gorm.io/gorm"
)

// Repository adalah interface/kontrak yang mendefinisikan apa saja yang bisa dilakukan ke database
type AuditRepository interface {
	CreateLog(log *models.AuditLog) error
	GetLogByHash(hash, clientID string) (*models.AuditLog, error)
	GetProofsByHash(hash string) ([]models.MerkleProof, error)
	GetDashboardStats(clientID string) (map[string]int64, error)
	GetLatestLogByResource(resource, clientID string) (*models.AuditLog, error)
	GetRecentLogs(limit int, clientID string) ([]models.AuditLog, error)
	GetResourceInventory(clientID string) ([]models.AuditLog, error)
	GetLogsByResource(resource, clientID string) ([]models.AuditLog, error)
}

// auditRepoImpl adalah implementasi nyata dari interface di atas menggunakan GORM
type auditRepoImpl struct {
	db *gorm.DB
}

// NewAuditRepository adalah fungsi pembuat (constructor)
func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepoImpl{db: db}
}

func (r *auditRepoImpl) CreateLog(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditRepoImpl) GetLogByHash(hash, clientID string) (*models.AuditLog, error) {
	var log models.AuditLog
	err := r.db.Where("hash_value = ? AND client_id = ?", hash, clientID).First(&log).Error
	return &log, err
}

func (r *auditRepoImpl) GetProofsByHash(hash string) ([]models.MerkleProof, error) {
	var proofs []models.MerkleProof
	// Ambil bukti (sibling hash) dan urutkan berdasarkan level pohon dari bawah ke atas
	err := r.db.Where("transaction_hash = ?", hash).Order("tree_level asc").Find(&proofs).Error
	return proofs, err
}

func (r *auditRepoImpl) GetDashboardStats(clientID string) (map[string]int64, error) {
	var totalLogs, anchoredLogs, pendingLogs int64

	r.db.Model(&models.AuditLog{}).Where("client_id = ?", clientID).Count(&totalLogs)
	r.db.Model(&models.AuditLog{}).Where("client_id = ? AND status = ?", clientID, "ANCHORED").Count(&anchoredLogs)
	r.db.Model(&models.AuditLog{}).Where("client_id = ? AND status IN ?", clientID, []string{"RECEIVED", "HASHED", "AGGREGATED"}).Count(&pendingLogs)

	return map[string]int64{
		"total_logs":    totalLogs,
		"anchored_logs": anchoredLogs,
		"pending_logs":  pendingLogs,
	}, nil
}

// GetLatestLogByResource mencari jejak log paling baru untuk suatu spesifik resource/data
func (r *auditRepoImpl) GetLatestLogByResource(resource, clientID string) (*models.AuditLog, error) {
	var log models.AuditLog
	// Urutkan berdasarkan waktu terbaru (descending) dan ambil yang pertama
	err := r.db.Where("resource = ? AND client_id = ?", resource, clientID).Order("timestamp desc").First(&log).Error
	return &log, err
}

func (r *auditRepoImpl) GetRecentLogs(limit int, clientID string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	// Mengambil log terbaru berdasarkan waktu, dibatasi sesuai parameter limit
	err := r.db.Where("client_id = ?", clientID).Order("timestamp desc").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *auditRepoImpl) GetResourceInventory(clientID string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	// Mengambil baris log terbaru untuk setiap resource unik
	err := r.db.Raw("SELECT DISTINCT ON (resource) * FROM audit_logs WHERE client_id = ? ORDER BY resource, timestamp DESC", clientID).Scan(&logs).Error
	return logs, err
}

func (r *auditRepoImpl) GetLogsByResource(resource, clientID string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	// Ambil dari yang terlama (ASC) untuk mengurutkan rantai PreviousHash
	err := r.db.Where("resource = ? AND client_id = ?", resource, clientID).Order("timestamp asc").Find(&logs).Error
	return logs, err
}
