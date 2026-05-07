package services

import (
	"auth-services/helpers"
	"auth-services/models"
	"auth-services/repositories"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthServiceInterface interface {
	IssueToken(ctx context.Context, merchantID, merchantSecret string) (*TokenResult, error)
}

type AuthService struct {
	merchantRepository repositories.MerchantRepositoryInterface
	tokenRepository    repositories.TokenRepositoryInterface
	jwtSecret          string
	jwtTTL             time.Duration
}

type TokenResult struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int64
}

func NewAuthService(merchantRepository repositories.MerchantRepositoryInterface, tokenRepository repositories.TokenRepositoryInterface, jwtSecret string, jwtTTLMinutes int64) *AuthService {
	return &AuthService{
		merchantRepository: merchantRepository,
		tokenRepository:    tokenRepository,
		jwtSecret:          jwtSecret,
		jwtTTL:             time.Duration(jwtTTLMinutes) * time.Minute,
	}
}

var ErrInvalidCredentials = errors.New("invalid credentials")

func (service *AuthService) IssueToken(ctx context.Context, merchantID, merchantSecret string) (*TokenResult, error) {
	merchant, err := service.merchantRepository.FindByMerchantID(ctx, merchantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !merchant.IsActive {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(merchant.SecretHash), []byte(merchantSecret)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, tokenID, expiresAt, err := helpers.GenerateJWT(service.jwtSecret, merchant.MerchantID, service.jwtTTL)
	if err != nil {
		return nil, err
	}

	accessToken := &models.AccessToken{
		ID:         uuid.New(),
		MerchantID: merchant.MerchantID,
		TokenID:    tokenID,
		ExpiresAt:  expiresAt,
	}

	if err := service.tokenRepository.Create(ctx, accessToken); err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(service.jwtTTL.Seconds()),
	}, nil
}
