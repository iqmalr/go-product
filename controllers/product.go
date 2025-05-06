package controllers

import (
	"go-product-api/events"
	"go-product-api/models"
	"go-product-api/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetProducts godoc
// @Summary Get all products
// @Description Get list of all products from Elasticsearch
// @Tags products
// @Produce json
// @Success 200 {array} models.Product
// @Router /products [get]
func GetProducts(c *gin.Context) {
	esRepo := repositories.NewElasticsearchRepository()
	products, err := esRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetProduct godoc
// @Summary Get product by ID
// @Description Get product details by product ID from Elasticsearch
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} models.Product
// @Failure 404 {object} object "Product not found"
// @Router /products/{id} [get]
func GetProduct(c *gin.Context) {
	esRepo := repositories.NewElasticsearchRepository()
	products, err := esRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

// CreateProduct godoc
// @Summary Create new product
// @Description Create a new product entry in PostgreSQL and send event to Kafka
// @Tags products
// @Accept json
// @Produce json
// @Param product body models.Product true "Product data"
// @Success 201 {object} models.Product
// @Failure 400 {object} object "Invalid input"
// @Router /products [post]
func CreateProduct(c *gin.Context) {
	var input models.Product
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pgRepo := repositories.NewPostgresRepository()
	if err := pgRepo.Create(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product: " + err.Error()})
		return
	}

	if err := events.PublishProductEvent(events.ProductCreated, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product created but failed to publish event: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

// UpdateProduct godoc
// @Summary Update product
// @Description Update existing product by ID in PostgreSQL and send event to Kafka
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body models.Product true "Updated product data"
// @Success 200 {object} models.Product
// @Failure 400 {object} object "Invalid input"
// @Failure 404 {object} object "Product not found"
// @Router /products/{id} [put]
func UpdateProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	pgRepo := repositories.NewPostgresRepository()
	product, err := pgRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var input models.Product
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.Name = input.Name
	product.Description = input.Description
	product.Price = input.Price

	if err := pgRepo.Update(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product: " + err.Error()})
		return
	}

	if err := events.PublishProductEvent(events.ProductUpdated, product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product updated but failed to publish event: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary Delete product
// @Description Delete product by ID from PostgreSQL and send event to Kafka
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} object "message: Product deleted"
// @Failure 404 {object} object "Product not found"
// @Router /products/{id} [delete]
func DeleteProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	pgRepo := repositories.NewPostgresRepository()
	product, err := pgRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if err := pgRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product: " + err.Error()})
		return
	}

	if err := events.PublishProductEvent(events.ProductDeleted, product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product deleted but failed to publish event: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}
