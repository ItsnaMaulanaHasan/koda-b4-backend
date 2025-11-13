package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func categoriesRoutes(r *gin.RouterGroup) {
	categories := r.Group("/categories")
	{
		categories.GET("", controllers.GetAllCategory)
		categories.GET("/:id", controllers.GetCategoryById)
		categories.POST("", controllers.CreateCategory)
		categories.PATCH("/:id", controllers.UpdateCategory)
		categories.DELETE("/:id", controllers.DeleteCategory)
	}
}
