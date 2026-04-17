package client

import (
	"go-blockchain-api/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	CreateClient(client *models.Client) error
}

type clientRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &clientRepository{db: db}
}

func (r *clientRepository) CreateClient(client *models.Client) error {
	return r.db.Create(client).Error
}
