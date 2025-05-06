package events

import (
	"encoding/json"
	"fmt"
	"go-product-api/config"
	"go-product-api/repositories"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func StartConsumer() {
	esRepo := repositories.NewElasticsearchRepository()

	err := config.KafkaConsumer.Subscribe(config.ProductTopic, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic %s: %v", config.ProductTopic, err)
	}

	go func() {
		for {
			msg, err := config.KafkaConsumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				log.Printf("Consumer error: %v", err)
				continue
			}
			if err := processMessage(msg, esRepo); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}
	}()
	log.Println("kafka consumer started")
}
func processMessage(msg *kafka.Message, esRepo *repositories.ElasticsearchRepository) error {
	var event ProductEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("error unmarshaling event: %w", err)
	}

	log.Printf("Processing %s event for product ID: %s", event.Type, event.Product.ID)

	switch event.Type {
	case ProductCreated, ProductUpdated:
		if err := esRepo.Index(event.Product); err != nil {
			return fmt.Errorf("error indexing product: %w", err)
		}
		log.Printf("Product indexed in Elasticsearch: %s", event.Product.ID)

	case ProductDeleted:
		if err := esRepo.Delete(event.Product.ID); err != nil {
			return fmt.Errorf("error deleting product: %w", err)
		}
		log.Printf("Product deleted from Elasticsearch: %s", event.Product.ID)

	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}

	return nil
}
