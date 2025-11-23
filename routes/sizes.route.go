package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func sizesRoutes(r *gin.Engine) {
	r.GET("/sizes", controllers.ListSizes)
}
