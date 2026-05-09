package repositories

import (
	"context"

	"payment-services/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repository *TransactionRepository) UpsertByRequestID(ctx context.Context, transaction *models.Transaction) error {
	return repository.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "request_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"customer_pan",
			"amount",
			"transaction_datetime",
			"rrn",
			"bill_number",
			"customer_name",
			"merchant_id",
			"merchant_name",
			"merchant_city",
			"currency_code",
			"payment_status",
			"payment_description",
			"updated_at",
		}),
	}).Create(transaction).Error
}

func (repository *TransactionRepository) FindByMerchantIDAndBillNumber(ctx context.Context, merchantID, billNumber string) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := repository.db.WithContext(ctx).Where("merchant_id = ? AND bill_number = ?", merchantID, billNumber).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}
