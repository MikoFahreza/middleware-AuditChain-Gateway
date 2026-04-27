package client

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{Service: service}
}

// Tambahkan field-field baru ke dalam struct DTO
type CreateClientRequest struct {
	CompanyName      string `json:"company_name" binding:"required" example:"PT Karya Bangsa"`
	SubscriptionTier string `json:"subscription_tier" example:"enterprise"`
	RateLimitPerSec  int    `json:"rate_limit_per_sec" example:"100"`
	Status           string `json:"status" example:"active"`
	ActorField       string `json:"actor_field" example:"operator_id"`
	ActionField      string `json:"action_field" example:"spatial_operation"`
	ResourceField    string `json:"resource_field" example:"layer_polygon"`
}

func (h *Handler) CreateClient(c *gin.Context) {
	var req CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tidak valid atau company_name belum diisi"})
		return
	}

	// Kita mengirim seluruh objek req (bukan cuma CompanyName) ke Service
	clientData, rawAPIKey, err := h.Service.RegisterClient(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Klien / Perusahaan SaaS berhasil didaftarkan",
		"client_id": clientData.ID,
		"api_key":   rawAPIKey,
	})
}