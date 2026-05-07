package helpers

import "github.com/gin-gonic/gin"

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSONError(ctx *gin.Context, statusCode int, code, message string) {
	ctx.JSON(statusCode, ErrorResponse{
		Code:    code,
		Message: message,
	})
}
