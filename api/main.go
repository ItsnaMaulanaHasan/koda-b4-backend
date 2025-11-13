package api

import (
	"backend-daily-greens/lib"
	"backend-daily-greens/routes"
	"net/http"

	_ "backend-daily-greens/docs"

	"github.com/gin-gonic/gin"
)

var App *gin.Engine

func init() {
	App := gin.Default()

	router := App.Group("/")

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, lib.ResponseSuccess{
			Success: true,
			Message: "Backend is running well",
		})
	})

	routes.SetUpRoutes(App)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	App.ServeHTTP(w, r)
}
