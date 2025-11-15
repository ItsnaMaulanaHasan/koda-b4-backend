package middlewares

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckSessionActive() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionId, exists := ctx.Get("sessionId")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
				Success: false,
				Message: "Session Id not found in token",
			})
			ctx.Abort()
			return
		}

		isActive, err := models.IsSessionActive(sessionId.(int))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to check session status",
				Error:   err.Error(),
			})
			ctx.Abort()
			return
		}

		if !isActive {
			ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
				Success: false,
				Message: "Session has been logged out",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
