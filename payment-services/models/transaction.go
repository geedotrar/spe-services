package models

import "time"

type Transaction struct {
	ID                  uint64    `gorm:"primaryKey"`
	RequestID           string    `gorm:"column:request_id"`
	CustomerPAN         string    `gorm:"column:customer_pan"`
	Amount              float64   `gorm:"column:amount"`
	TransactionDatetime time.Time `gorm:"column:transaction_datetime"`
	RRN                 string    `gorm:"column:rrn"`
	BillNumber          string    `gorm:"column:bill_number"`
	CustomerName        *string   `gorm:"column:customer_name"`
	MerchantID          string    `gorm:"column:merchant_id"`
	MerchantName        string    `gorm:"column:merchant_name"`
	MerchantCity        string    `gorm:"column:merchant_city"`
	CurrencyCode        string    `gorm:"column:currency_code"`
	PaymentStatus       string    `gorm:"column:payment_status"`
	PaymentDescription  *string   `gorm:"column:payment_description"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func (Transaction) TableName() string {
	return "transactions"
}
