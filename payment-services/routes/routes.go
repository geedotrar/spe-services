package routes

import (
	"payment-services/handlers"
	"payment-services/middlewares"
	"payment-services/services"

	"github.com/gin-gonic/gin"
)

func Register(engine *gin.Engine, paymentHandler *handlers.PaymentHandler, paymentService *services.PaymentService, jwtSecret string) {
	api := engine.Group("/api/v1")
	api.Use(middlewares.AuthMiddleware(jwtSecret))
	api.POST("/transaction-notification", middlewares.NotificationSignatureMiddleware(paymentService), paymentHandler.TransactionNotification)
	api.POST("/check-status", middlewares.CheckStatusSignatureMiddleware(paymentService), paymentHandler.CheckStatus)

	// helper endpoint for generate signature for dev
	dev := engine.Group("/dev")
	dev.POST("/signature", paymentHandler.GenerateSignature)
}
