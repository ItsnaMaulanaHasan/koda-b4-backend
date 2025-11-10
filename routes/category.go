package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func SetupCategoryRoutes(r *gin.Engine, admin *gin.RouterGroup) {
	categories := admin.Group("/categories")
	{
		categories.GET("", controllers.GetAllCategory)
		categories.GET("/:id", controllers.GetCategoryById)
		categories.POST("", controllers.CreateCategory)
		categories.PATCH("/:id", controllers.UpdateCategory)
	}
}
