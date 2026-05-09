package middlewares

import (
	"net/http"
	"strings"

	"payment-services/helpers"
	"payment-services/services"

	"github.com/gin-gonic/gin"
)

func NotificationSignatureMiddleware(paymentService *services.PaymentService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request struct {
			RequestID  string `json:"request_id"`
			RRN        string `json:"rrn"`
			MerchantID string `json:"merchant_id"`
		}

		if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
			helpers.JSONError(ctx, http.StatusBadRequest, "01", "invalid request payload")
			ctx.Abort()
			return
		}

		merchant, err := paymentService.GetMerchant(ctx.Request.Context(), request.MerchantID)
		if err != nil {
			helpers.JSONError(ctx, http.StatusUnauthorized, "04", "merchant is not registered")
			ctx.Abort()
			return
		}

		expectedSignature := helpers.SignHMACSHA512(strings.Join([]string{request.RequestID, request.RRN, request.MerchantID}, ":"), merchant.SigningKey)
		if ctx.GetHeader("X-Signature") != expectedSignature {
			helpers.JSONError(ctx, http.StatusUnauthorized, "05", "invalid signature")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func CheckStatusSignatureMiddleware(paymentService *services.PaymentService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request struct {
			BillNumber string `json:"bill_number"`
		}

		if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
			helpers.JSONError(ctx, http.StatusBadRequest, "01", "invalid request payload")
			ctx.Abort()
			return
		}

		merchantID := GetMerchantID(ctx)
		merchant, err := paymentService.GetMerchant(ctx.Request.Context(), merchantID)
		if err != nil {
			helpers.JSONError(ctx, http.StatusUnauthorized, "04", "merchant is not registered")
			ctx.Abort()
			return
		}

		expectedSignature := helpers.SignHMACSHA512(request.BillNumber, merchant.SigningKey)
		if ctx.GetHeader("X-Signature") != expectedSignature {
			helpers.JSONError(ctx, http.StatusUnauthorized, "05", "invalid signature")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
