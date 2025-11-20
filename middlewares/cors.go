package middlewares

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{os.Getenv("ORIGIN_URL"), os.Getenv("ORIGIN_URL_VERCEL"), "http://localhost:5173"},
		AllowMethods:     []string{"PATCH", "POST", "GET", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	})
}
