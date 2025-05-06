package config

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

var ES *elasticsearch.Client

func ConnectElasticsearch() {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.Info(client.Info.WithContext(ctx))
	if err != nil {
		log.Fatalf("Error getting Elasticsearch info: %s", err)
	}
	defer res.Body.Close()

	ES = client
	fmt.Println("Elasticsearch connection established")

	createProductIndex()
}

func createProductIndex() {
	mapping := `{
		"mappings": {
			"properties": {
				"id": { "type": "keyword" },
				"name": { "type": "text" },
				"description": { "type": "text" },
				"price": { "type": "integer" }
			}
		}
	}`

	res, err := ES.Indices.Exists([]string{"products"})
	if err != nil {
		log.Fatalf("Error checking if index exists: %s", err)
	}

	if res.StatusCode == 404 {
		res, err := ES.Indices.Create(
			"products",
			ES.Indices.Create.WithBody(strings.NewReader(mapping)),
		)
		if err != nil {
			log.Fatalf("Error creating index: %s", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			log.Fatalf("Error creating index: %s", res.String())
		}
		
		fmt.Println("Products index created successfully")
	}
}