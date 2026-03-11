package hasher

import (
	"fmt"
	"go-blockchain-api/internal/models"
	"go-blockchain-api/pkg/crypto"
	"log"

	"gorm.io/gorm"
)

type HasherEngine struct {
	DB *gorm.DB
}

// ProcessPendingLogs mencari log berstatus RECEIVED dan memprosesnya menjadi HASHED [cite: 164, 175]
func (h *HasherEngine) ProcessPendingLogs() error {
	var pendingLogs []models.AuditLog

	// 1. Ambil log yang belum di-hash, urutkan berdasarkan waktu agar rantai hash teratur [cite: 173]
	if err := h.DB.Where("status = ?", "RECEIVED").Order("timestamp asc").Find(&pendingLogs).Error; err != nil {
		return err
	}

	if len(pendingLogs) == 0 {
		return nil // Tidak ada log baru yang perlu diproses
	}

	for _, auditLog := range pendingLogs {
		// 2. Ambil Hash dari transaksi terakhir untuk dijadikan Previous Hash [cite: 168]
		var lastLog models.AuditLog
		var prevHash string

		result := h.DB.Where("status IN ?", []string{"HASHED", "ANCHORED"}).Order("timestamp desc").First(&lastLog)
		if result.Error == nil {
			prevHash = lastLog.HashValue
		} else {
			// Jika ini adalah log pertama di database (Genesis)
			prevHash = "GENESIS_00000000000000000000000000000000000000000000000000000000"
		}

		// 3. Serialisasi Data Secara Deterministik (Urutan tidak boleh berubah) [cite: 166-167, 171-173]
		// Menggabungkan: log_id | actor | action | resource | timestamp | source_system | auth_context | previous_hash | metadata
		contextString := fmt.Sprintf("%s|%s|%s|%s|%d|%s|%s|%s|%s",
			auditLog.LogID,
			auditLog.Actor,
			auditLog.Action,
			auditLog.Resource,
			auditLog.Timestamp.UnixNano(), // Gunakan Nano agar presisi
			auditLog.SourceSystem,
			auditLog.AuthorizationContext,
			prevHash,
			auditLog.Metadata,
		)

		// 4. Generate SHA3-256 Hash
		hashValue := crypto.GenerateSHA3_256(contextString)

		// 5. Update Record di Database [cite: 170]
		auditLog.HashValue = hashValue
		auditLog.PreviousHash = prevHash
		auditLog.Status = "HASHED"

		if err := h.DB.Save(&auditLog).Error; err != nil {
			log.Printf("[Hasher] Gagal menyimpan hash untuk log %s: %v", auditLog.LogID, err)
			continue
		}

		log.Printf("[Hasher] ✅ Log %s berhasil di-hash: %s", auditLog.LogID, hashValue)
	}

	return nil
}
