package repositories

import (
	"auth-services/models"
	"context"

	"gorm.io/gorm"
)

type MerchantRepositoryInterface interface {
	FindByMerchantID(ctx context.Context, merchantID string) (*models.Merchant, error)
}

type MerchantRepository struct {
	db *gorm.DB
}

func NewMerchantRepository(db *gorm.DB) *MerchantRepository {
	return &MerchantRepository{db: db}
}

func (repository *MerchantRepository) FindByMerchantID(ctx context.Context, merchantID string) (*models.Merchant, error) {
	var merchant models.Merchant
	if err := repository.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&merchant).Error; err != nil {
		return nil, err
	}
	return &merchant, nil
}
