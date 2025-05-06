package repositories

import (
	"go-product-api/config"
	"go-product-api/models"

	"github.com/google/uuid"
)

type PostgresRepository struct{}

func NewPostgresRepository() *PostgresRepository {
	return &PostgresRepository{}
}

func (r *PostgresRepository) FindAll() ([]models.Product, error) {
	var products []models.Product
	result := config.DB.Find(&products)
	return products, result.Error
}

func (r *PostgresRepository) FindByID(id uuid.UUID) (models.Product, error) {
	var product models.Product
	result := config.DB.First(&product, "id = ?", id)
	return product, result.Error
}

func (r *PostgresRepository) Create(product *models.Product) error {
	return config.DB.Create(product).Error
}

func (r *PostgresRepository) Update(product *models.Product) error {
	return config.DB.Save(product).Error
}

func (r *PostgresRepository) Delete(id uuid.UUID) error {
	return config.DB.Delete(&models.Product{}, "id = ?", id).Error
}
