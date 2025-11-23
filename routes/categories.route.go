package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func categoriesRoutes(r *gin.Engine, admin *gin.RouterGroup) {
	categories := admin.Group("/categories")
	{
		categories.GET("/:id", controllers.DetailCategory)
		categories.POST("", controllers.CreateCategory)
		categories.PATCH("/:id", controllers.UpdateCategory)
		categories.DELETE("/:id", controllers.DeleteCategory)
	}

	r.GET("/categories", controllers.ListCategories)
}
