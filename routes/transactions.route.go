package routes

import (
	"backend-daily-greens/controllers"
	"backend-daily-greens/middlewares"

	"github.com/gin-gonic/gin"
)

func transactionsRoutes(r *gin.Engine, admin *gin.RouterGroup) {
	transactions := admin.Group("/transactions")
	{
		transactions.GET("", controllers.ListTransactions)
		transactions.GET("/:id", controllers.DetailTransactions)
		transactions.PATCH("/:id", controllers.UpdateTransactionStatus)
	}

	r.POST("/transactions", middlewares.Auth(), controllers.Checkout)
}
