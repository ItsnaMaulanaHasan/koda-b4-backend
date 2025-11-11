package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(r *gin.Engine, admin *gin.RouterGroup) {
	users := admin.Group("/users")
	{
		users.GET("", controllers.GetAllUser)
		users.GET("/:id", controllers.GetUserById)
		users.POST("", controllers.CreateUser)
		users.PATCH("/:id", controllers.UpdateUser)
		users.DELETE("/:id", controllers.DeleteUser)
	}
}
