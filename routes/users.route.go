package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func usersRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		users.GET("", controllers.ListUsers)
		users.GET("/:id", controllers.DetailUser)
		users.POST("", controllers.CreateUser)
		users.PATCH("/:id", controllers.UpdateUser)
		users.DELETE("/:id", controllers.DeleteUser)
	}
}
