package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func profilesRoutes(r *gin.RouterGroup) {
	r.GET("", controllers.DetailProfile)
	// r.PATCH("", controllers.UpdateProfile)
	// r.PATCH("", controllers.UploadPhotoProfile)
}
