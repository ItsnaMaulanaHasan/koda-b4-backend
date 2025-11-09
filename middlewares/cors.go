package middlewares

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func CorsMiddleware() gin.HandlerFunc {
	godotenv.Load()
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", os.Getenv("ORIGIN_URL"))
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type")
		ctx.Next()
	}
}
