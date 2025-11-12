package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func usersRoutes(r *gin.Engine, admin *gin.RouterGroup) {
	users := admin.Group("/users")
	{
		users.GET("", controllers.ListUsers)
		users.GET("/:id", controllers.DetailUser)
		users.POST("", controllers.CreateUser)
		users.PATCH("/:id", controllers.UpdateUser)
		users.DELETE("/:id", controllers.DeleteUser)
	}
}
