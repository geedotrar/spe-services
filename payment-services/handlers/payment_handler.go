package handlers

import (
	"errors"
	"net/http"
	"payment-services/helpers"
	"payment-services/middlewares"
	"payment-services/services"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type PaymentHandlerInterface interface {
	ProcessPayment(ctx *gin.Context)
}

type PaymentHandler struct {
	paymentService *services.PaymentService
}

type NotificationRequest struct {
	RequestID           string  `json:"request_id" binding:"required,max=32"`
	CustomerPAN         string  `json:"customer_pan" binding:"required,max=32"`
	Amount              float64 `json:"amount" binding:"required"`
	TransactionDatetime string  `json:"transaction_datetime" binding:"required"`
	RRN                 string  `json:"rrn" binding:"required,max=32"`
	BillNumber          string  `json:"bill_number" binding:"required,max=64"`
	CustomerName        *string `json:"customer_name"`
	MerchantID          string  `json:"merchant_id" binding:"required,max=32"`
	MerchantName        string  `json:"merchant_name" binding:"required"`
	MerchantCity        string  `json:"merchant_city" binding:"required"`
	CurrencyCode        string  `json:"currency_code" binding:"required,max=8"`
	PaymentStatus       string  `json:"payment_status" binding:"required,max=8"`
	PaymentDescription  *string `json:"payment_description"`
}

type CheckStatusRequest struct {
	RequestID  string `json:"request_id" binding:"required,max=32"`
	BillNumber string `json:"bill_number" binding:"required,max=64"`
}

type BaseResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CheckStatusResponse struct {
	Code                string  `json:"code"`
	Message             string  `json:"message"`
	RequestID           string  `json:"request_id"`
	CustomerPAN         string  `json:"customer_pan"`
	Amount              float64 `json:"amount"`
	TransactionDatetime string  `json:"transaction_datetime"`
	RRN                 string  `json:"rrn"`
	BillNumber          string  `json:"bill_number"`
	CustomerName        *string `json:"customer_name,omitempty"`
	MerchantID          string  `json:"merchant_id"`
	MerchantName        string  `json:"merchant_name"`
	MerchantCity        string  `json:"merchant_city"`
	CurrencyCode        string  `json:"currency_code"`
	PaymentStatus       string  `json:"payment_status"`
	PaymentDescription  *string `json:"payment_description,omitempty"`
}

func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

func (handler *PaymentHandler) TransactionNotification(ctx *gin.Context) {
	var request NotificationRequest
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		helpers.JSONError(ctx, http.StatusBadRequest, "01", "invalid request payload")
		return
	}

	if merchantID := middlewares.GetMerchantID(ctx); merchantID != request.MerchantID {
		helpers.JSONError(ctx, http.StatusUnauthorized, "03", "merchant token mismatch")
		return
	}

	transactionTime, err := time.Parse(time.RFC3339, request.TransactionDatetime)
	if err != nil {
		helpers.JSONError(ctx, http.StatusBadRequest, "01", "invalid transaction_datetime format")
		return
	}

	err = handler.paymentService.ProcessNotification(ctx.Request.Context(), services.NotificationInput{
		RequestID:           request.RequestID,
		CustomerPAN:         request.CustomerPAN,
		Amount:              request.Amount,
		TransactionDatetime: transactionTime,
		RRN:                 request.RRN,
		BillNumber:          request.BillNumber,
		CustomerName:        request.CustomerName,
		MerchantID:          request.MerchantID,
		MerchantName:        request.MerchantName,
		MerchantCity:        request.MerchantCity,
		CurrencyCode:        request.CurrencyCode, PaymentStatus: request.PaymentStatus,
		PaymentDescription: request.PaymentDescription,
	})
	if err != nil {
		if errors.Is(err, services.ErrDuplicateBillNumber) {
			helpers.JSONError(ctx, http.StatusConflict, "04", "bill_number already used by another transaction")
			return
		}
		helpers.JSONError(ctx, http.StatusInternalServerError, "99", "internal server error")
		return
	}

	ctx.JSON(http.StatusOK, BaseResponse{Code: "00", Message: "success"})
}

func (handler *PaymentHandler) CheckStatus(ctx *gin.Context) {
	var request CheckStatusRequest
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		helpers.JSONError(ctx, http.StatusBadRequest, "01", "invalid request payload")
		return
	}

	transaction, err := handler.paymentService.CheckStatus(ctx.Request.Context(), services.CheckStatusInput{
		MerchantID: middlewares.GetMerchantID(ctx),
		BillNumber: request.BillNumber,
	})
	if err != nil {
		if errors.Is(err, services.ErrTransactionNotFound) {
			helpers.JSONError(ctx, http.StatusNotFound, "14", "transaction not found")
			return
		}

		helpers.JSONError(ctx, http.StatusInternalServerError, "99", "internal server error")
		return
	}

	ctx.JSON(http.StatusOK, CheckStatusResponse{
		Code:                "00",
		Message:             "success",
		RequestID:           transaction.RequestID,
		CustomerPAN:         transaction.CustomerPAN,
		Amount:              transaction.Amount,
		TransactionDatetime: transaction.TransactionDatetime.UTC().Format(time.RFC3339),
		RRN:                 transaction.RRN,
		BillNumber:          transaction.BillNumber,
		CustomerName:        transaction.CustomerName,
		MerchantID:          transaction.MerchantID,
		MerchantName:        transaction.MerchantName,
		MerchantCity:        transaction.MerchantCity,
		CurrencyCode:        transaction.CurrencyCode,
		PaymentStatus:       transaction.PaymentStatus,
		PaymentDescription:  transaction.PaymentDescription,
	})
}

func (handler *PaymentHandler) GenerateSignature(ctx *gin.Context) {
	var request struct {
		RequestID  string `json:"request_id"`
		RRN        string `json:"rrn"`
		MerchantID string `json:"merchant_id"`
		BillNumber string `json:"bill_number"`
		Key        string `json:"key" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		helpers.JSONError(ctx, http.StatusBadRequest, "01", "field 'key' wajib diisi")
		return
	}

	parts := make([]string, 0, 4)
	if request.RequestID != "" {
		parts = append(parts, request.RequestID)
	}
	if request.RRN != "" {
		parts = append(parts, request.RRN)
	}
	if request.MerchantID != "" {
		parts = append(parts, request.MerchantID)
	}
	if request.BillNumber != "" {
		parts = append(parts, request.BillNumber)
	}

	if len(parts) == 0 {
		helpers.JSONError(ctx, http.StatusBadRequest, "01", "fill min 1 (request_id, rrn, merchant_id, atau bill_number)")
		return
	}

	payload := strings.Join(parts, ":")
	signature := helpers.SignHMACSHA512(payload, request.Key)

	ctx.JSON(http.StatusOK, gin.H{
		"code":      "00",
		"message":   "success",
		"payload":   payload,
		"signature": signature,
	})
}
