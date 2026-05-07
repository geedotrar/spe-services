package repositories

import (
	"auth-services/models"
	"context"

	"gorm.io/gorm"
)

type TokenRepositoryInterface interface {
	Create(ctx context.Context, token *models.AccessToken) error
}

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (repository *TokenRepository) Create(ctx context.Context, token *models.AccessToken) error {
	return repository.db.WithContext(ctx).Create(token).Error
}
