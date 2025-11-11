package routes

import (
	"backend-daily-greens/controllers"

	"github.com/gin-gonic/gin"
)

func featuresRouter(r *gin.Engine) {
	r.GET("/favourite-products", controllers.ListFavouriteProducts)
}
