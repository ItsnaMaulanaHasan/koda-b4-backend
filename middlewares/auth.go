package middlewares

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.Request.Header.Get("Authorization")
		tokenString, found := strings.CutPrefix(authHeader, "Bearer ")
		if !found {
			ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
				Success: false,
				Message: "Authorization header required or invalid format",
			})
			ctx.Abort()
			return
		}

		blacklistKey := "blacklist:" + tokenString
		exists, err := config.Rdb.Exists(context.Background(), blacklistKey).Result()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, lib.ResponseError{
				Success: false,
				Message: "Failed to verify token",
				Error:   err.Error(),
			})
			ctx.Abort()
			return
		}

		if exists > 0 {
			ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
				Success: false,
				Message: "Token has been revoked, please login again",
			})
			ctx.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &lib.UserPayload{}, func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv("APP_SECRET")), nil
		})
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
				Success: false,
				Message: "Invalid or expired token",
			})
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(*lib.UserPayload)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
				Success: false,
				Message: "Invalid token claims",
			})
			ctx.Abort()
			return
		}

		ctx.Set("userId", claims.Id)
		ctx.Set("role", claims.Role)

		ctx.Next()
	}
}
