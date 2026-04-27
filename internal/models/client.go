package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Client merepresentasikan perusahaan/tenant penyewa layanan
type Client struct {
	ID          string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CompanyName string `gorm:"type:varchar(100);not null" json:"company_name"`

	// Keamanan API Key
	APIKeyPrefix string `gorm:"type:varchar(20);uniqueIndex;not null" json:"-"`
	APIKeyHash   string `gorm:"type:varchar(255);not null" json:"-"`

	// Konfigurasi SaaS
	SubscriptionTier string `gorm:"type:varchar(20);default:'basic'" json:"subscription_tier"`
	RateLimitPerSec  int    `gorm:"default:50" json:"rate_limit_per_sec"`
	Status           string `gorm:"type:varchar(20);default:'active'" json:"status"`

	// 👇 TAMBAHAN BARU: Konfigurasi Mapping Field Dinamis
	ActorField    string `gorm:"type:varchar(100)" json:"actor_field"`
	ActionField   string `gorm:"type:varchar(100)" json:"action_field"`
	ResourceField string `gorm:"type:varchar(100)" json:"resource_field"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relasi (Has Many)
	Users     []User     `gorm:"foreignKey:ClientID" json:"-"`
	AuditLogs []AuditLog `gorm:"foreignKey:ClientID" json:"-"`
}

func (c *Client) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return
}