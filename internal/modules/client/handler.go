package client

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"go-blockchain-api/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

type CreateClientRequest struct {
	CompanyName string `json:"company_name" binding:"required" example:"Rumah Sakit Sentosa"`
}

func generateRandomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CreateClient mendaftarkan perusahaan dan membuat API Key
// @Summary Mendaftarkan Klien SaaS Baru
// @Description Menambahkan perusahaan dan mencetak API Key rahasia (Hanya muncul 1x)
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body CreateClientRequest true "Nama Perusahaan"
// @Success 201 {object} map[string]interface{} "Klien berhasil dibuat"
// @Router /admin/clients [post]
func (h *Handler) CreateClient(c *gin.Context) {
	var input CreateClientRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama perusahaan wajib diisi"})
		return
	}

	// 1. Generate API Key (Misal: ak_live_ + 16 karakter hex acak)
	randomSecret := generateRandomHex(8) // 8 bytes = 16 char hex
	fullAPIKey := "ak_live_" + randomSecret
	prefix := "ak_live_" + randomSecret[:5]

	// 2. Hash API Key untuk disimpan di database
	hash := sha256.Sum256([]byte(fullAPIKey))
	hashedKey := hex.EncodeToString(hash[:])

	newClient := models.Client{
		CompanyName:  input.CompanyName,
		APIKeyPrefix: prefix,
		APIKeyHash:   hashedKey,
		Status:       "active",
	}

	if err := h.DB.Create(&newClient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan klien ke database"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Klien berhasil didaftarkan",
		"client_id":    newClient.ID,
		"company_name": newClient.CompanyName,
		"api_key":      fullAPIKey,
		"warning":      "SIMPAN API KEY INI SEKARANG! Sistem tidak akan menampilkannya lagi.",
	})
}
