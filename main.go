package main

// @title           API Documentation
// @version         1.0
// @description     Dokumentasi REST API menggunakan Gin dan Swagger

// @host      localhost:8080
// @BasePath  /

import (
	"backend-daily-greens/config"
	"backend-daily-greens/middlewares"
	"backend-daily-greens/routes"

	_ "backend-daily-greens/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	godotenv.Load()
	config.ConnectDatabase()
	defer config.CloseDatabase()

	r := gin.Default()

	r.MaxMultipartMemory = 1 << 20

	r.Use(middlewares.AllowPrefic())
	r.Use(middlewares.CorsMiddleware())

	routes.SetUpRoutes(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8080")
}
