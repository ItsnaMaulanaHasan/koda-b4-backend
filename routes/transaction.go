package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func SetupTransactionRoutes(r *gin.Engine, admin *gin.RouterGroup) {
	transactions := admin.Group("/transactions")
	{
		transactions.GET("", controllers.GetAllTransaction)
		transactions.GET("/:id", controllers.GetTransactionById)
	}
}
