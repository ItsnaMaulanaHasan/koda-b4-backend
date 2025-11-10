package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func SetupProductRoutes(r *gin.Engine, admin *gin.RouterGroup) {
	products := admin.Group("/products")
	{
		products.GET("", controllers.GetAllProduct)
		products.GET("/:id", controllers.GetProductById)
	}
}
