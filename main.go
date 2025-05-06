package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-product-api/config"
	_ "go-product-api/docs"
	"go-product-api/events"
	"go-product-api/models"
	"go-product-api/routes"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

// @title           Product API
// @version         1.0
// @description     REST API sederhana dengan Golang dan PostgreSQL.
// @host            localhost:8082
// @BasePath        /
func main() {
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	config.ConnectDatabase()
	config.DB.AutoMigrate(&models.Product{})
	config.ConnectElasticsearch()
	config.ConnectKafka()
	defer config.CloseKafkaConnections()
	events.StartConsumer()
	routes.SetupRoutes(r)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		log.Println("Shutting down gracefully...")
		config.CloseKafkaConnections()
		os.Exit(0)
	}()

	r.Run(":8082")
}
