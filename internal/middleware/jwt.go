package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth adalah middleware untuk melindungi endpoint Dashboard
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Cek apakah ada header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak. Token tidak ditemukan."})
			c.Abort() // Menghentikan request agar tidak lanjut ke Handler
			return
		}

		// 2. Format token harus "Bearer <token_string>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Format token salah. Gunakan format: Bearer <token>"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		secret := os.Getenv("JWT_SECRET")

		// 3. Validasi keaslian dan masa berlaku token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Pastikan algoritma yang digunakan adalah HMAC (HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrAbortHandler
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token tidak valid atau sudah kedaluwarsa. Silakan login kembali.",
			})
			c.Abort()
			return
		}

		// Jika token sah, persilakan tamu masuk ke ruangan (Handler)
		c.Next()
	}
}
