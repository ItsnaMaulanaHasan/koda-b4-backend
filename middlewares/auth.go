package middlewares

import (
	"backend-daily-greens/lib"
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

		token, err := jwt.ParseWithClaims(tokenString, &lib.UserPayload{}, func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv("APP_SECRET")), nil
		})

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
				Success: false,
				Message: "Invalid token",
			})
			ctx.Abort()
			return
		}

		if claims, ok := token.Claims.(*lib.UserPayload); ok && token.Valid {
			ctx.Set("userId", claims.Id)
			ctx.Next()
		} else {
			ctx.JSON(http.StatusUnauthorized, lib.ResponseError{
				Success: false,
				Message: "Invalid token claims",
			})
			ctx.Abort()
			return
		}
	}
}
