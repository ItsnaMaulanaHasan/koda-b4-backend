package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func categoriesRoutes(r *gin.RouterGroup) {
	categories := r.Group("/categories")
	{
		categories.GET("", controllers.ListCategories)
		categories.GET("/:id", controllers.DetailCategory)
		categories.POST("", controllers.CreateCategory)
		categories.PATCH("/:id", controllers.UpdateCategory)
		categories.DELETE("/:id", controllers.DeleteCategory)
	}
}
