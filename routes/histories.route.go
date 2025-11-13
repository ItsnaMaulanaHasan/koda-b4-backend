package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func historiesRoutes(r *gin.RouterGroup) {
	r.GET("", controllers.ListHistories)
	r.GET("/:id", controllers.DetailHistoriy)
}
