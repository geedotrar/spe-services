package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type JSONB []byte

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

func (j *JSONB) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*j = nil
		return nil
	case []byte:
		*j = append((*j)[:0], v...)
		return nil
	case string:
		*j = append((*j)[:0], v...)
		return nil
	default:
		return fmt.Errorf("unsupported JSONB scan type %T", value)
	}
}

type TransactionEventLog struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	RequestID    string    `gorm:"column:request_id" json:"request_id"`
	MerchantID   string    `gorm:"column:merchant_id" json:"merchant_id"`
	BillNumber   string    `gorm:"column:bill_number" json:"bill_number"`
	Status       string    `gorm:"column:status" json:"status"`
	EventPayload JSONB     `gorm:"column:event_payload;type:jsonb" json:"event_payload"`
	PublishedAt  time.Time `gorm:"column:published_at" json:"published_at"`
	ConsumedAt   time.Time `gorm:"column:consumed_at" json:"consumed_at"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

func (TransactionEventLog) TableName() string {
	return "transaction_event_logs"
}

func (tel *TransactionEventLog) ParseEventPayload(v interface{}) error {
	if tel.EventPayload == nil || len(tel.EventPayload) == 0 {
		return nil
	}
	return json.Unmarshal(tel.EventPayload, v)
}
