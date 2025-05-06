package events

import (
	"encoding/json"
	"fmt"
	"go-product-api/config"
	"go-product-api/models"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type EventType string

const (
	ProductCreated EventType = "product_created"
	ProductUpdated EventType = "product_updated"
	ProductDeleted EventType = "product_deleted"
)

type ProductEvent struct {
	Type    EventType      `json:"type"`
	Product models.Product `json:"product"`
}

func PublishProductEvent(eventType EventType, product models.Product) error {
	event := ProductEvent{
		Type:    eventType,
		Product: product,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling product event: %w", err)
	}

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &config.ProductTopic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(product.ID.String()),
		Value: payload,
	}

	if err := config.KafkaProducer.Produce(message, nil); err != nil {
		return fmt.Errorf("error publishing to Kafka: %w", err)
	}

	log.Printf("Published event: %s for product ID: %s\n", eventType, product.ID)
	return nil
}
