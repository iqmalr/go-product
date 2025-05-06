package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-product-api/config"
	"go-product-api/models"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
)

type ElasticsearchRepository struct{}

func NewElasticsearchRepository() *ElasticsearchRepository {
	return &ElasticsearchRepository{}
}

func (r *ElasticsearchRepository) FindAll() ([]models.Product, error) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"size": 100,
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding query: %s", err)
	}

	res, err := config.ES.Search(
		config.ES.Search.WithContext(context.Background()),
		config.ES.Search.WithIndex("products"),
		config.ES.Search.WithBody(&buf),
		config.ES.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}

	var products []models.Product
	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})

		product := models.Product{
			ID:          uuid.MustParse(source["id"].(string)),
			Name:        source["name"].(string),
			Description: source["description"].(string),
			Price:       int(source["price"].(float64)),
		}

		products = append(products, product)
	}

	return products, nil
}

func (r *ElasticsearchRepository) FindByID(id uuid.UUID) (models.Product, error) {
	req := esapi.GetRequest{
		Index:      "products",
		DocumentID: id.String(),
	}

	res, err := req.Do(context.Background(), config.ES)
	if err != nil {
		return models.Product{}, fmt.Errorf("error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return models.Product{}, errors.New("product not found")
	}

	if res.IsError() {
		return models.Product{}, fmt.Errorf("error response: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return models.Product{}, fmt.Errorf("error parsing response body: %s", err)
	}

	source := result["_source"].(map[string]interface{})
	product := models.Product{
		ID:          uuid.MustParse(source["id"].(string)),
		Name:        source["name"].(string),
		Description: source["description"].(string),
		Price:       int(source["price"].(float64)),
	}

	return product, nil
}

func (r *ElasticsearchRepository) Index(product models.Product) error {

	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      "products",
		DocumentID: product.ID.String(),
		Body:       strings.NewReader(string(productJSON)),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), config.ES)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index error: %s", res.String())
	}

	return nil
}

func (r *ElasticsearchRepository) Delete(id uuid.UUID) error {
	req := esapi.DeleteRequest{
		Index:      "products",
		DocumentID: id.String(),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), config.ES)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete error: %s", res.String())
	}

	return nil
}
