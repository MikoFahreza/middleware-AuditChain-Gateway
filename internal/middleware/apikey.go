package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth adalah middleware untuk melindungi rute Machine-to-Machine (M2M)
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil kunci rahasia dari environment variable
		expectedKey := os.Getenv("INGESTION_API_KEY")
		if expectedKey == "" {
			// Fail-safe: Jika admin lupa set .env, blokir semua akses agar tidak bocor
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server misconfiguration: API Key not set"})
			c.Abort()
			return
		}

		// Ambil API Key yang dikirim oleh klien melalui Header
		clientKey := c.GetHeader("api-key")

		// Validasi
		if clientKey == "" || clientKey != expectedKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Akses Ditolak: API Key tidak valid atau tidak ditemukan di Header 'X-API-Key'",
			})
			c.Abort() // Hentikan request di sini, jangan teruskan ke Handler
			return
		}

		// Jika lolos, izinkan request lanjut ke Handler (ReceiveLog)
		c.Next()
	}
}
