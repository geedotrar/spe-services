package repositories

import (
	"context"

	"payment-services/models"

	"gorm.io/gorm"
)

type TransactionEventLogRepository struct {
	db *gorm.DB
}

func NewTransactionEventLogRepository(db *gorm.DB) *TransactionEventLogRepository {
	return &TransactionEventLogRepository{db: db}
}

func (repo *TransactionEventLogRepository) CreateEventLog(ctx context.Context, eventLog *models.TransactionEventLog) error {
	return repo.db.WithContext(ctx).Create(eventLog).Error
}
