package routes

import (
	"backend-daily-greens/middlewares"

	"github.com/gin-gonic/gin"
)

func SetUpRoutes(r *gin.Engine) {
	SetupAuthRoutes(r)

	admin := r.Group("/admin", middlewares.Auth(), middlewares.AdminOnly())
	SetupUserRoutes(r, admin)
	SetupProductRoutes(r, admin)
	SetupTransactionRoutes(r, admin)
	SetupCategoryRoutes(r, admin)
}
