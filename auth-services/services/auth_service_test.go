package services

import (
	"context"
	"testing"
	"time"

	models "auth-services/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// --- mock MerchantRepository ---
type mockMerchantRepo struct {
	mock.Mock
}

// --- mock TokenRepository ---
type mockTokenRepo struct {
	mock.Mock
}

func (m *mockMerchantRepo) FindByMerchantID(ctx context.Context, merchantID string) (*models.Merchant, error) {
	args := m.Called(ctx, merchantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Merchant), args.Error(1)
}

func (m *mockTokenRepo) Create(ctx context.Context, token *models.AccessToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func makeActivateMerchant(merchantID, plainSecret string) *models.Merchant {
	hash, _ := bcrypt.GenerateFromPassword([]byte(plainSecret), bcrypt.MinCost)
	return &models.Merchant{
		ID:         1,
		MerchantID: merchantID,
		Name:       "Test",
		SecretHash: string(hash),
		SigningKey: "signing-key",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// --- test cases ---
func TestTokenValidLogin(t *testing.T) {
	ctx := context.Background()
	merchantID := "123467890"
	plainSecret := "merchant-secret"

	merchant := makeActivateMerchant(merchantID, plainSecret)

	merchantRepo := new(mockMerchantRepo)
	tokenRepo := new(mockTokenRepo)

	merchantRepo.On("FindByMerchantID", ctx, merchantID).Return(merchant, nil)
	tokenRepo.On("Create", ctx, mock.AnythingOfType("*models.AccessToken")).Return(nil)

	svc := NewAuthService(merchantRepo, tokenRepo, "jwt-secret", 60)
	result, err := svc.IssueToken(ctx, merchantID, plainSecret)

	require.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, int64(3600), result.ExpiresIn)
	merchantRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

func TestTokenInvalidLogin(t *testing.T) {
	ctx := context.Background()

	merchantRepo := new(mockMerchantRepo)
	tokenRepo := new(mockTokenRepo)

	merchantRepo.On("FindByMerchantID", ctx, "MERCHANT-NOT-EXIST").
		Return(nil, gorm.ErrRecordNotFound)

	svc := NewAuthService(merchantRepo, tokenRepo, "jwt-secret", 60)
	result, err := svc.IssueToken(ctx, "MERCHANT-NOT-EXIST", "any-secret")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestTokenInvalidMerchantInactive(t *testing.T) {
	ctx := context.Background()
	merchantID := "M0000001"

	merchant := makeActivateMerchant(merchantID, "secret")
	merchant.IsActive = false // merchant not active

	merchantRepo := new(mockMerchantRepo)
	tokenRepo := new(mockTokenRepo)

	merchantRepo.On("FindByMerchantID", ctx, merchantID).Return(merchant, nil)

	svc := NewAuthService(merchantRepo, tokenRepo, "jwt-secret", 60)
	result, err := svc.IssueToken(ctx, merchantID, "secret")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestTokenInvalidSecret(t *testing.T) {
	ctx := context.Background()
	merchantID := "M0000001"

	merchant := makeActivateMerchant(merchantID, "true-secret")

	merchantRepo := new(mockMerchantRepo)
	tokenRepo := new(mockTokenRepo)

	merchantRepo.On("FindByMerchantID", ctx, merchantID).Return(merchant, nil)

	svc := NewAuthService(merchantRepo, tokenRepo, "jwt-secret", 60)
	result, err := svc.IssueToken(ctx, merchantID, "wrong-secret")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestTokenStoredWithCorrectMerchantID(t *testing.T) {
	ctx := context.Background()
	merchantID := "M0000001"
	plainSecret := "secret"

	merchant := makeActivateMerchant(merchantID, plainSecret)

	merchantRepo := new(mockMerchantRepo)
	tokenRepo := new(mockTokenRepo)

	merchantRepo.On("FindByMerchantID", ctx, merchantID).Return(merchant, nil)

	var savedToken *models.AccessToken
	tokenRepo.On("Create", ctx, mock.MatchedBy(func(t *models.AccessToken) bool {
		savedToken = t
		return true
	})).Return(nil)

	svc := NewAuthService(merchantRepo, tokenRepo, "jwt-secret", 60)
	_, err := svc.IssueToken(ctx, merchantID, plainSecret)

	require.NoError(t, err)
	require.NotNil(t, savedToken)
	assert.Equal(t, merchantID, savedToken.MerchantID)
	assert.NotEqual(t, uuid.Nil, savedToken.ID)
}
