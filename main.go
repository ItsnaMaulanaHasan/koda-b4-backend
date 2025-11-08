package main

// @title           API Documentation
// @version         1.0
// @description     Dokumentasi REST API menggunakan Gin dan Swagger

// @host      localhost:8080
// @BasePath  /

import (
	"backend-daily-greens/config"
	"backend-daily-greens/routes"

	_ "backend-daily-greens/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	config.ConnectDatabase()
	defer config.CloseDatabase()

	r := gin.Default()

	routes.SetUpRoutes(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8080")
}
