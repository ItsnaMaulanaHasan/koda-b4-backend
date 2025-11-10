package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func SetupCategoryRoutes(r *gin.Engine, admin *gin.RouterGroup) {
	categories := admin.Group("/categories")
	{
		categories.GET("", controllers.GetAllCategory)
	}
}
