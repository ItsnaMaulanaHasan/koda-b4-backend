package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(r *gin.Engine) {
	users := r.Group("/users")
	{
		users.GET("", controllers.GetAllUser)
		users.GET("/:id", controllers.GetUserById)
	}
}
