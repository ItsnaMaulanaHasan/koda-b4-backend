package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
		auth.POST("/forgot-password", controllers.GetTokenReset)
		auth.POST("/verify-reset-token", controllers.VerifyResetToken)
		auth.POST("/reset-password", controllers.ResetPassword)
	}
}
