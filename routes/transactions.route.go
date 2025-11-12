package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func transactionsRoutes(r *gin.Engine, admin *gin.RouterGroup) {
	transactions := admin.Group("/transactions")
	{
		transactions.GET("", controllers.GetAllTransaction)
		transactions.GET("/:id", controllers.GetTransactionById)
		transactions.PATCH("/:id", controllers.UpdateTransactionStatus)
	}
}
