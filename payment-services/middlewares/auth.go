package middlewares

import (
	"net/http"

	"payment-services/helpers"

	"github.com/gin-gonic/gin"
)

const merchantIDContextKey = "merchant_id"

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, err := helpers.ParseBearerToken(jwtSecret, ctx.GetHeader("Authorization"))
		if err != nil {
			helpers.JSONError(ctx, http.StatusUnauthorized, "03", "invalid bearer token")
			ctx.Abort()
			return
		}

		ctx.Set(merchantIDContextKey, claims.MerchantID)
		ctx.Next()
	}
}

func GetMerchantID(ctx *gin.Context) string {
	value, ok := ctx.Get(merchantIDContextKey)
	if !ok {
		return ""
	}

	merchantID, _ := value.(string)
	return merchantID
}
