package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	KafkaProducer *kafka.Producer
	KafkaConsumer *kafka.Consumer
)

var (
	ProductTopic = "product_events"
)

func ConnectKafka() {
	// Use localhost:29092 which is the PLAINTEXT_HOST listener
	bootstrapServers := "localhost:29092"

	// Producer configuration
	producerConfig := kafka.ConfigMap{
		"bootstrap.servers":       bootstrapServers,
		"client.id":               "go-product-api",
		"socket.keepalive.enable": true,
	}

	producer, err := kafka.NewProducer(&producerConfig)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %s", err)
	}

	// Consumer configuration
	consumerConfig := kafka.ConfigMap{
		"bootstrap.servers":  bootstrapServers,
		"group.id":           "go-product-group",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
		"session.timeout.ms": 10000,
		"socket.timeout.ms":  30000,
	}

	consumer, err := kafka.NewConsumer(&consumerConfig)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %s", err)
	}

	KafkaProducer = producer
	KafkaConsumer = consumer

	fmt.Println("Kafka connection established")

	// Start a goroutine to handle delivery reports
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition.Error)
				} else {
					log.Printf("Message delivered to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	// Try to create the topic
	ensureTopicExists()
}

func ensureTopicExists() {
	// Create admin client
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:29092",
	})
	if err != nil {
		log.Printf("Failed to create admin client: %v\n", err)
		return
	}
	defer adminClient.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the topic
	topics := []kafka.TopicSpecification{
		{
			Topic:             ProductTopic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	// Try to create topic
	results, err := adminClient.CreateTopics(ctx, topics)
	if err != nil {
		log.Printf("Failed to create topics: %v\n", err)
		return
	}

	// Check results
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError &&
			result.Error.Code() != kafka.ErrTopicAlreadyExists {
			log.Printf("Failed to create topic %s: %v\n", result.Topic, result.Error)
		} else {
			log.Printf("Topic %s created or already exists\n", result.Topic)
		}
	}

	// Test topic metadata retrieval to confirm connection
	metadata, err := adminClient.GetMetadata(nil, true, 10000)
	if err != nil {
		log.Printf("Failed to get metadata: %v\n", err)
		return
	}

	log.Printf("Connected to Kafka cluster with %d brokers\n", len(metadata.Brokers))
	for _, broker := range metadata.Brokers {
		log.Printf("Broker: %d at %s\n", broker.ID, broker.Host)
	}
}

func CloseKafkaConnections() {
	if KafkaProducer != nil {
		KafkaProducer.Flush(15 * 1000) // Wait up to 15 seconds for messages to be delivered
		KafkaProducer.Close()
	}
	if KafkaConsumer != nil {
		KafkaConsumer.Close()
	}
}
