package models

import "time"

type Merchant struct {
	ID         uint64    `gorm:"primaryKey"`
	MerchantID string    `gorm:"column:merchant_id"`
	Name       string    `gorm:"column:name"`
	SecretHash string    `gorm:"column:secret_hash"`
	SigningKey string    `gorm:"column:signing_key"`
	IsActive   bool      `gorm:"column:is_active"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (Merchant) TableName() string {
	return "merchants"
}
