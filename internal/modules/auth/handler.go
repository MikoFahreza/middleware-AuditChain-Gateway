package auth

import (
	"net/http"
	"os"
	"time"

	"go-blockchain-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

// Struct khusus untuk Login (Tidak perlu client_id, cukup username & password)
type AuthRequest struct {
	Username string `json:"username" binding:"required" example:"auditor_senior"`
	Password string `json:"password" binding:"required,min=6" example:"rahasia1234"`
}

// Struct khusus untuk Register (Wajib menyertakan client_id/Perusahaan)
type RegisterRequest struct {
	ClientID string `json:"client_id" binding:"required" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
	Username string `json:"username" binding:"required" example:"auditor_senior"`
	Password string `json:"password" binding:"required,min=6" example:"rahasia1234"`
}

// Register mendaftarkan akun baru ke dalam sistem perusahaan.
// @Summary Pendaftaran Akun Auditor
// @Description Mendaftarkan pengguna baru dan mengaitkannya dengan perusahaan (client_id).
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Kredensial Pengguna Baru & ID Perusahaan"
// @Success 201 {object} map[string]interface{} "Akun berhasil dibuat"
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tidak valid atau client_id belum diisi."})
		return
	}

	// 1. Validasi: Apakah ClientID tersebut ada di database?
	var client models.Client
	if err := h.DB.First(&client, "id = ?", req.ClientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Perusahaan (Client ID) tidak terdaftar di sistem"})
		return
	}

	// 2. Hash password menggunakan Bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses kata sandi"})
		return
	}

	// 3. Buat objek User baru (Ikat dengan ClientID)
	newUser := models.User{
		ClientID: req.ClientID, // 👈 Pengikat Multi-Tenant
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     "Auditor",
	}

	// 4. Simpan ke database
	if err := h.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username sudah digunakan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pengguna berhasil didaftarkan ke perusahaan " + client.CompanyName,
		"user": map[string]interface{}{
			"id":        newUser.ID,
			"client_id": newUser.ClientID,
			"username":  newUser.Username,
			"role":      newUser.Role,
		},
	})
}

// Login memverifikasi user dan mencetak JWT dengan Data Perusahaan
// @Summary Login Auditor
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body AuthRequest true "Kredensial Login"
// @Success 200 {object} map[string]interface{} "Berhasil Login"
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid"})
		return
	}

	var user models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau Password salah!"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau Password salah!"})
		return
	}

	// TAMBAHAN SAAS: Masukkan client_id ke dalam JWT!
	// Ini krusial agar Dashboard tahu data perusahaan mana yang boleh ditampilkan
	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"client_id": user.ClientID, // 👈 Diselipkan ke dalam identitas JWT
		"username":  user.Username,
		"role":      user.Role,
		"exp":       time.Now().Add(time.Hour * 2).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencetak token keamanan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"token":   tokenString,
	})
}
