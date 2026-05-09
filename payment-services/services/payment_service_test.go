package services

import (
	"context"
	"testing"
	"time"

	"payment-services/models"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockMerchantRepo struct {
	mock.Mock
}

func (m *mockMerchantRepo) FindByMerchantID(ctx context.Context, merchantID string) (*models.Merchant, error) {
	args := m.Called(ctx, merchantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Merchant), args.Error(1)
}

type mockTransactionRepo struct {
	mock.Mock
}

func (m *mockTransactionRepo) UpsertByRequestID(ctx context.Context, transaction *models.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *mockTransactionRepo) FindByMerchantIDAndBillNumber(ctx context.Context, merchantID, billNumber string) (*models.Transaction, error) {
	args := m.Called(ctx, merchantID, billNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func newTestService(merchantRepo merchantRepository, txRepo transactionRepository) *PaymentService {
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	return NewPaymentService(merchantRepo, txRepo, redisClient, nil, 30*time.Second)
}

func makeTransaction() *models.Transaction {
	name := "John Doe"
	desc := "Payment Success"
	return &models.Transaction{
		RequestID:           "REQ-001",
		CustomerPAN:         "9360001110000000019",
		Amount:              10000.00,
		TransactionDatetime: time.Now(),
		RRN:                 "123456789012",
		BillNumber:          "BILL-001",
		CustomerName:        &name,
		MerchantID:          "008800223497",
		MerchantName:        "Sukses Makmur",
		MerchantCity:        "Jakarta",
		CurrencyCode:        "360",
		PaymentStatus:       "00",
		PaymentDescription:  &desc,
	}
}

func TestProcessNotification_Success(t *testing.T) {
	ctx := context.Background()
	tx := makeTransaction()

	txRepo := new(mockTransactionRepo)
	txRepo.On("UpsertByRequestID", ctx, mock.AnythingOfType("*models.Transaction")).Return(nil)

	svc := newTestService(new(mockMerchantRepo), txRepo)
	err := svc.ProcessNotification(ctx, NotificationInput{
		RequestID:           tx.RequestID,
		CustomerPAN:         tx.CustomerPAN,
		Amount:              tx.Amount,
		TransactionDatetime: tx.TransactionDatetime,
		RRN:                 tx.RRN,
		BillNumber:          tx.BillNumber,
		CustomerName:        tx.CustomerName,
		MerchantID:          tx.MerchantID,
		MerchantName:        tx.MerchantName,
		MerchantCity:        tx.MerchantCity,
		CurrencyCode:        tx.CurrencyCode,
		PaymentStatus:       tx.PaymentStatus,
		PaymentDescription:  tx.PaymentDescription,
	})

	require.NoError(t, err)
	txRepo.AssertExpectations(t)
}

func TestProcessNotification_DuplicateBillNumber(t *testing.T) {
	ctx := context.Background()
	tx := makeTransaction()

	pgErr := &pgconn.PgError{Code: "23505", ConstraintName: "transactions_bill_number_key"}

	txRepo := new(mockTransactionRepo)
	txRepo.On("UpsertByRequestID", ctx, mock.AnythingOfType("*models.Transaction")).Return(pgErr)

	svc := newTestService(new(mockMerchantRepo), txRepo)
	err := svc.ProcessNotification(ctx, NotificationInput{
		RequestID:           tx.RequestID,
		CustomerPAN:         tx.CustomerPAN,
		Amount:              tx.Amount,
		TransactionDatetime: tx.TransactionDatetime,
		RRN:                 tx.RRN,
		BillNumber:          tx.BillNumber,
		MerchantID:          tx.MerchantID,
		MerchantName:        tx.MerchantName,
		MerchantCity:        tx.MerchantCity,
		CurrencyCode:        tx.CurrencyCode,
		PaymentStatus:       tx.PaymentStatus,
	})

	assert.ErrorIs(t, err, ErrDuplicateBillNumber)
}

func TestProcessNotification_DBError(t *testing.T) {
	ctx := context.Background()
	tx := makeTransaction()

	txRepo := new(mockTransactionRepo)
	txRepo.On("UpsertByRequestID", ctx, mock.AnythingOfType("*models.Transaction")).Return(gorm.ErrInvalidDB)

	svc := newTestService(new(mockMerchantRepo), txRepo)
	err := svc.ProcessNotification(ctx, NotificationInput{
		RequestID:           tx.RequestID,
		MerchantID:          tx.MerchantID,
		BillNumber:          tx.BillNumber,
		MerchantName:        tx.MerchantName,
		MerchantCity:        tx.MerchantCity,
		CurrencyCode:        tx.CurrencyCode,
		PaymentStatus:       tx.PaymentStatus,
		TransactionDatetime: tx.TransactionDatetime,
	})

	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrDuplicateBillNumber)
}

func TestCheckStatus_Success(t *testing.T) {
	ctx := context.Background()
	tx := makeTransaction()

	txRepo := new(mockTransactionRepo)
	txRepo.On("FindByMerchantIDAndBillNumber", ctx, tx.MerchantID, tx.BillNumber).Return(tx, nil)

	svc := newTestService(new(mockMerchantRepo), txRepo)
	result, err := svc.CheckStatus(ctx, CheckStatusInput{
		MerchantID: tx.MerchantID,
		BillNumber: tx.BillNumber,
	})

	require.NoError(t, err)
	assert.Equal(t, tx.RequestID, result.RequestID)
	assert.Equal(t, tx.PaymentStatus, result.PaymentStatus)
	txRepo.AssertExpectations(t)
}

func TestCheckStatus_NotFound(t *testing.T) {
	ctx := context.Background()

	txRepo := new(mockTransactionRepo)
	txRepo.On("FindByMerchantIDAndBillNumber", ctx, "M001", "BILL-999").
		Return(nil, gorm.ErrRecordNotFound)

	svc := newTestService(new(mockMerchantRepo), txRepo)
	result, err := svc.CheckStatus(ctx, CheckStatusInput{
		MerchantID: "M001",
		BillNumber: "BILL-999",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrTransactionNotFound)
}

func TestCheckStatus_DBError(t *testing.T) {
	ctx := context.Background()

	txRepo := new(mockTransactionRepo)
	txRepo.On("FindByMerchantIDAndBillNumber", ctx, "M001", "BILL-001").
		Return(nil, gorm.ErrInvalidDB)

	svc := newTestService(new(mockMerchantRepo), txRepo)
	result, err := svc.CheckStatus(ctx, CheckStatusInput{
		MerchantID: "M001",
		BillNumber: "BILL-001",
	})

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrTransactionNotFound)
}

func TestGetMerchant_Success(t *testing.T) {
	ctx := context.Background()
	merchant := &models.Merchant{
		MerchantID: "M001",
		Name:       "Test Merchant",
		SigningKey: "speskilltest",
		IsActive:   true,
	}

	merchantRepo := new(mockMerchantRepo)
	merchantRepo.On("FindByMerchantID", ctx, "M001").Return(merchant, nil)

	svc := newTestService(merchantRepo, new(mockTransactionRepo))
	result, err := svc.GetMerchant(ctx, "M001")

	require.NoError(t, err)
	assert.Equal(t, "M001", result.MerchantID)
	merchantRepo.AssertExpectations(t)
}

func TestGetMerchant_NotFound(t *testing.T) {
	ctx := context.Background()

	merchantRepo := new(mockMerchantRepo)
	merchantRepo.On("FindByMerchantID", ctx, "UNKNOWN").Return(nil, gorm.ErrRecordNotFound)

	svc := newTestService(merchantRepo, new(mockTransactionRepo))
	result, err := svc.GetMerchant(ctx, "UNKNOWN")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrMerchantNotFound)
}
