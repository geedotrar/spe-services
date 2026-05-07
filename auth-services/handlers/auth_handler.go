package handlers

import (
	"auth-services/helpers"
	"auth-services/services"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandlerInterface interface {
	IssueToken(ctx *gin.Context)
}

type AuthHandler struct {
	authService services.AuthServiceInterface
}

type TokenRequest struct {
	MerchantID     string `json:"merchant_id" binding:"required,max=32"`
	MerchantSecret string `json:"merchant_secret" binding:"required"`
}

type TokenResponse struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (handler *AuthHandler) IssueToken(ctx *gin.Context) {
	var request TokenRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		helpers.JSONError(ctx, http.StatusBadRequest, "01", "invalid request payload")
		return
	}

	result, err := handler.authService.IssueToken(ctx.Request.Context(), request.MerchantID, request.MerchantSecret)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			helpers.JSONError(ctx, http.StatusUnauthorized, "02", "invalid merchant credentials")
			return
		}

		helpers.JSONError(ctx, http.StatusInternalServerError, "99", "internal server error")
		return
	}

	ctx.JSON(http.StatusOK, TokenResponse{
		Code:        "00",
		Message:     "success",
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		ExpiresIn:   result.ExpiresIn,
	})
}
