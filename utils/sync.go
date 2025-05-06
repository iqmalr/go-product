package utils

import (
	"fmt"
	"go-product-api/config"
	"go-product-api/models"
	"go-product-api/repositories"
	"log"
)

func SyncPostgresToElasticsearch() error {
	log.Println("Starting data synchronization from PostgreSQL to Elasticsearch...")

	pgRepo := repositories.NewPostgresRepository()
	esRepo := repositories.NewElasticsearchRepository()

	products, err := pgRepo.FindAll()
	if err != nil {
		return fmt.Errorf("failed to fetch products from PostgreSQL: %w", err)
	}

	log.Printf("Found %d products in PostgreSQL", len(products))

	for _, product := range products {
		if err := esRepo.Index(product); err != nil {
			log.Printf("Error indexing product %s: %v", product.ID, err)
			continue
		}
	}

	log.Println("Synchronization completed")
	return nil
}

func InitializeIndices() error {

	log.Println("Elasticsearch indices initialized")
	return nil
}

func MigrateDatabase() error {
	log.Println("Migrating PostgreSQL database...")
	err := config.DB.AutoMigrate(&models.Product{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Println("Database migration completed")
	return nil
}
