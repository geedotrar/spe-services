package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"payment-services/models"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ErrMerchantNotFound = errors.New("merchant not found")
var ErrTransactionNotFound = errors.New("transaction not found")
var ErrDuplicateBillNumber = errors.New("duplicate bill_number")

type merchantRepository interface {
	FindByMerchantID(ctx context.Context, merchantID string) (*models.Merchant, error)
}

type transactionRepository interface {
	UpsertByRequestID(ctx context.Context, transaction *models.Transaction) error
	FindByMerchantIDAndBillNumber(ctx context.Context, merchantID, billNumber string) (*models.Transaction, error)
}

type NotificationInput struct {
	RequestID           string
	CustomerPAN         string
	Amount              float64
	TransactionDatetime time.Time
	RRN                 string
	BillNumber          string
	CustomerName        *string
	MerchantID          string
	MerchantName        string
	MerchantCity        string
	CurrencyCode        string
	PaymentStatus       string
	PaymentDescription  *string
}

type CheckStatusInput struct {
	MerchantID string
	BillNumber string
}

type PaymentService struct {
	merchantRepository    merchantRepository
	transactionRepository transactionRepository
	redisClient           *redis.Client
	eventService          *EventService
	checkStatusCacheTTL   time.Duration
}

func NewPaymentService(merchantRepository merchantRepository, transactionRepository transactionRepository, redisClient *redis.Client, eventService *EventService, checkStatusCacheTTL time.Duration) *PaymentService {
	return &PaymentService{
		merchantRepository:    merchantRepository,
		transactionRepository: transactionRepository,
		redisClient:           redisClient,
		eventService:          eventService,
		checkStatusCacheTTL:   checkStatusCacheTTL,
	}
}

func (service *PaymentService) GetMerchant(ctx context.Context, merchantID string) (*models.Merchant, error) {
	merchant, err := service.merchantRepository.FindByMerchantID(ctx, merchantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMerchantNotFound
		}
		return nil, err
	}
	return merchant, nil
}

func (service *PaymentService) ProcessNotification(ctx context.Context, input NotificationInput) error {
	transaction := &models.Transaction{
		RequestID:           input.RequestID,
		CustomerPAN:         input.CustomerPAN,
		Amount:              input.Amount,
		TransactionDatetime: input.TransactionDatetime,
		RRN:                 input.RRN,
		BillNumber:          input.BillNumber,
		CustomerName:        input.CustomerName,
		MerchantID:          input.MerchantID,
		MerchantName:        input.MerchantName,
		MerchantCity:        input.MerchantCity,
		CurrencyCode:        input.CurrencyCode,
		PaymentStatus:       input.PaymentStatus,
		PaymentDescription:  input.PaymentDescription,
		UpdatedAt:           time.Now().UTC(),
	}

	if err := service.transactionRepository.UpsertByRequestID(ctx, transaction); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "transactions_bill_number_key" {
			return ErrDuplicateBillNumber
		}
		return err
	}

	cacheKey := statusCacheKey(input.MerchantID, input.BillNumber)
	service.redisClient.Del(ctx, cacheKey)

	if service.eventService != nil {
		_ = service.eventService.PublishTransaction(ctx, TransactionEvent{
			RequestID:   input.RequestID,
			MerchantID:  input.MerchantID,
			BillNumber:  input.BillNumber,
			Status:      input.PaymentStatus,
			PublishedAt: time.Now().UTC(),
		})
	}

	return nil
}

func (service *PaymentService) CheckStatus(ctx context.Context, input CheckStatusInput) (*models.Transaction, error) {
	cacheKey := statusCacheKey(input.MerchantID, input.BillNumber)
	cachedPayload, err := service.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var transaction models.Transaction
		if unmarshalErr := json.Unmarshal([]byte(cachedPayload), &transaction); unmarshalErr == nil {
			return &transaction, nil
		}
	}

	transaction, err := service.transactionRepository.FindByMerchantIDAndBillNumber(ctx, input.MerchantID, input.BillNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}

	serialized, marshalErr := json.Marshal(transaction)
	if marshalErr == nil {
		service.redisClient.Set(ctx, cacheKey, string(serialized), service.checkStatusCacheTTL)
	}

	return transaction, nil
}

func statusCacheKey(merchantID, billNumber string) string {
	return fmt.Sprintf("payment:status:%s:%s", merchantID, billNumber)
}
