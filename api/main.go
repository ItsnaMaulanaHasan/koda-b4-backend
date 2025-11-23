package handler

import (
	"backend-daily-greens/config"
	"backend-daily-greens/lib"
	"backend-daily-greens/middlewares"
	"backend-daily-greens/routes"
	"fmt"
	"net/http"
	"os"
	"sync"

	_ "backend-daily-greens/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	App  *gin.Engine
	once sync.Once
)

func initApp() {
	once.Do(func() {
		fmt.Println("ORIGIN_URL:", os.Getenv("ORIGIN_URL"))
		fmt.Println("ORIGIN_URL_VERCEL:", os.Getenv("ORIGIN_URL_VERCEL"))
		fmt.Println("ORIGIN_URL_VERCEL2:", os.Getenv("ORIGIN_URL_VERCEL2"))

		App = gin.New()
		App.Use(gin.Recovery())
		App.Use(middlewares.CorsMiddleware())

		App.GET("/", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, lib.ResponseSuccess{
				Success: true,
				Message: "Backend is running well",
			})
		})

		App.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		App.MaxMultipartMemory = 1 << 20

		routes.SetUpRoutes(App)
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	initApp()

	config.InitDatabase()
	config.InitRedis()
	config.InitSupabase()

	App.ServeHTTP(w, r)
}
