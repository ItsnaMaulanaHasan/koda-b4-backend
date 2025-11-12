package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func cartsRouter(r *gin.RouterGroup) {
	r.GET("", controllers.ListCarts)
	r.POST("", controllers.AddCart)
	r.DELETE("/:id", controllers.DeleteCart)
}
