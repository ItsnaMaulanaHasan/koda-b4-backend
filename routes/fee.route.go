package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func feeRoutes(r *gin.Engine) {
	r.GET("/order-methods", controllers.GetAllOrderMethods)
	r.GET("/payment-methods", controllers.GetAllPaymentMethods)
}
