package models

import (
	"time"

	"github.com/google/uuid"
)

type AccessToken struct {
	ID         uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	MerchantID string    `gorm:"column:merchant_id"`
	TokenID    uuid.UUID `gorm:"column:token_id;type:uuid"`
	ExpiresAt  time.Time `gorm:"column:expires_at"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (AccessToken) TableName() string {
	return "access_tokens"
}
