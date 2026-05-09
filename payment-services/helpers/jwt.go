package helpers

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type MerchantClaims struct {
	MerchantID string `json:"merchant_id"`
	jwt.RegisteredClaims
}

func ParseBearerToken(secret, header string) (*MerchantClaims, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, errors.New("invalid authorization header")
	}

	token, err := jwt.ParseWithClaims(parts[1], &MerchantClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*MerchantClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
