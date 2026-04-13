package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	// Sesuaikan path import ini dengan lokasi model Anda
	"go-blockchain-api/internal/models"
)

// APIKeyAuth adalah middleware untuk melindungi rute Machine-to-Machine (M2M)
// Berubah: Sekarang menerima *gorm.DB sebagai parameter
func APIKeyAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil API Key yang dikirim oleh klien (Mendukung header 'x-api-key' atau 'api-key')
		clientKey := c.GetHeader("x-api-key")
		if clientKey == "" {
			clientKey = c.GetHeader("api-key")
		}

		// Validasi Format Minimal
		if clientKey == "" || len(clientKey) < 12 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Akses Ditolak: API Key tidak valid atau tidak ditemukan di Header",
			})
			c.Abort() // Hentikan request
			return
		}

		// Ekstrak Prefix dan Hash Key
		prefix := clientKey[:10]
		hashBytes := sha256.Sum256([]byte(clientKey))
		hashedKey := hex.EncodeToString(hashBytes[:])

		var client models.Client

		// Validasi Kunci ke Database PostgreSQL via GORM
		err := db.Where("api_key_prefix = ? AND api_key_hash = ? AND status = ?", prefix, hashedKey, "active").First(&client).Error
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Akses Ditolak: API Key tidak dikenali atau akun klien tidak aktif",
			})
			c.Abort()
			return
		}

		// Jika lolos, INJEKSIKAN ID Klien ke dalam Context agar bisa dipakai oleh Handler
		c.Set("client_id", client.ID)

		// Lanjutkan request ke Handler (ReceiveLog)
		c.Next()
	}
}
