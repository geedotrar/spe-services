package helpers

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateJWT_SuccessfullyGeneratesToken(t *testing.T) {
	secret := "test-secret"

	merchantID := "MERCHANT-001"

	ttl := 30 * time.Minute

	tokenStr, tokenID, expiresAt, err := GenerateJWT(secret, merchantID, ttl)

	require.NoError(t, err)

	assert.NotEmpty(t, tokenStr)

	assert.NotEqual(t, tokenID.String(), "00000000-0000-0000-0000-000000000000")

	assert.True(t, expiresAt.After(time.Now()))
}

func TestGenerateJWT_ClaimsAreCorrect(t *testing.T) {
	secret := "test-secret"

	merchantID := "MERCHANT-001"

	ttl := 15 * time.Minute

	tokenStr, tokenID, _, err := GenerateJWT(secret, merchantID, ttl)

	require.NoError(t, err)

	parsed, err := jwt.ParseWithClaims(
		tokenStr,
		&MerchantClaims{},

		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)

	require.NoError(t, err)

	claims, ok := parsed.Claims.(*MerchantClaims)

	require.True(t, ok)

	assert.Equal(t, merchantID, claims.MerchantID)

	assert.Equal(t, tokenID.String(), claims.ID)
}

func TestGenerateJWT_ExpirationMatchesTTL(t *testing.T) {

	secret := "test-secret"

	ttl := 5 * time.Minute

	_, _, expiresAt, err := GenerateJWT(secret, "M001", ttl)

	require.NoError(t, err)

	diff := expiresAt.Sub(time.Now())

	assert.True(t, diff > 4*time.Minute && diff < 6*time.Minute)
}
