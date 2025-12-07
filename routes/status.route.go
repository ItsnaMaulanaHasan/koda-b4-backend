package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func statusRoutes(r *gin.RouterGroup) {
	users := r.Group("/status")
	{
		users.GET("", controllers.ListStatus)
	}
}
