package helpers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type MerchantClaims struct {
	MerchantID string `json:"merchant_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(secret string, merchantID string, ttl time.Duration) (string, uuid.UUID, time.Time, error) {
	tokenID := uuid.New()
	expiresAt := time.Now().UTC().Add(ttl)

	claims := MerchantClaims{
		MerchantID: merchantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			Subject:   merchantID,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", uuid.Nil, time.Time{}, err
	}

	return signedToken, tokenID, expiresAt, nil
}
