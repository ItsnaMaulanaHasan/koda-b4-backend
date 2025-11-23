package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func variantsRoutes(r *gin.Engine) {
	r.GET("/variants", controllers.ListVariants)
}
