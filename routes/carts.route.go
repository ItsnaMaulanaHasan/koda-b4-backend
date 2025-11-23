package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func cartsRouter(r *gin.Engine, cart *gin.RouterGroup) {
	cart.POST("", controllers.AddCart)
	cart.DELETE("/:id", controllers.DeleteCart)

	r.GET("/carts", controllers.ListCarts)
}
