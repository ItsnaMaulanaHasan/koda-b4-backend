package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func authRouter(r *gin.RouterGroup) {
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	r.POST("/logout", controllers.Logout)
	r.POST("/forgot-password", controllers.GetTokenReset)
	r.POST("/verify-reset-token", controllers.VerifyResetToken)
	r.PATCH("/reset-password", controllers.ResetPassword)
}
