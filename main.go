// @title           Product API
// @version         1.0
// @description     REST API sederhana dengan Golang dan PostgreSQL.
// @host      localhost:8082
// @BasePath  /
package main

import (
	"go-product-api/config"
	_ "go-product-api/docs"
	"go-product-api/models"
	"go-product-api/routes"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	config.ConnectDatabase()
	config.DB.AutoMigrate(&models.Product{})

	routes.SetupRoutes(r)

	r.Run(":8082")
}