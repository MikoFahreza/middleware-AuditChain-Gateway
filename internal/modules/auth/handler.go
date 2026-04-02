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

// ---------------------------------------------------------
// INI DIA YANG DICARI OLEH GOLANG: Struct Handler (H Kapital)
// ---------------------------------------------------------
type Handler struct {
	DB *gorm.DB
}

type AuthRequest struct {
	Username string `json:"username" binding:"required" example:"auditor_senior"`
	Password string `json:"password" binding:"required,min=6" example:"rahasia1234"`
}

// Register mendaftarkan akun baru ke dalam sistem.
// @Summary Pendaftaran Akun
// @Description Mendaftarkan pengguna baru. Secara otomatis akan diberikan role "Auditor" oleh sistem.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body AuthRequest true "Kredensial Pengguna Baru"
// @Success 201 {object} map[string]interface{} "Akun berhasil dibuat"
// @Failure 400 {object} map[string]interface{} "Format input tidak valid (misal: password kurang dari 6 karakter)"
// @Failure 409 {object} map[string]interface{} "Username sudah terdaftar di database"
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tidak valid. Password minimal 6 karakter."})
		return
	}

	// 1. Hash password menggunakan Bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses kata sandi"})
		return
	}

	// 2. Buat objek User baru
	newUser := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     "Auditor", // Default role
	}

	// 3. Simpan ke database
	if err := h.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username sudah digunakan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pengguna berhasil didaftarkan",
		"user":    newUser,
	})
}

// Login memverifikasi user dari database dan mencetak JWT
// @Summary Login Auditor
// @Description Mengotentikasi user dan mengembalikan JWT Token untuk akses Dashboard
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body AuthRequest true "Kredensial Login"
// @Success 200 {object} map[string]interface{} "Berhasil Login"
// @Failure 400 {object} map[string]interface{} "Format tidak valid"
// @Failure 401 {object} map[string]interface{} "Username/Password salah"
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid"})
		return
	}

	// 1. Cari user di database
	var user models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau Password salah!"})
		return
	}

	// 2. Bandingkan password yang diketik dengan hash di database
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau Password salah!"})
		return
	}

	// 3. Jika cocok, buatkan JWT
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 2).Unix(),
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
