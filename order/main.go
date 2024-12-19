package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"github.com/tankcdr/ppe-kafka-go/common"
	kafka "github.com/tankcdr/ppe-kafka-go/common/kafka"
	order "github.com/tankcdr/ppe-kafka-go/order/pkg"
)

// Config holds the environment configuration
type Config struct {
	Broker string `env:"KAFKA_BROKER" envDefault:"localhost:29092"`
	Topic  string `env:"KAFKA_TOPIC" envDefault:"order-received"`
}

// AppDependencies holds shared dependencies like Kafka producers
type AppDependencies struct {
	Producer *kafka.KafkaProducer
	Config   Config
}

func main() {
	// Load configuration
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Kafka producer
	kafkaConfig := kafka.KafkaConfig{
		Brokers: []string{cfg.Broker},
		Topic:   cfg.Topic,
		GroupID: "order-service",
	}
	producer := kafka.NewProducer(kafkaConfig)
	defer producer.Close()

	// Create shared dependencies
	deps := AppDependencies{
		Producer: producer,
		Config:   cfg,
	}

	// Setup and run the server
	r := setupRouter(&deps)
	r.Run(":8080")
}

func setupRouter(deps *AppDependencies) *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})
	r.POST("/order", postOrder(deps))
	return r
}

// postOrder handles the POST /order route
// expects a JSON payload of an Order
func postOrder(deps *AppDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		// validating the JSON payload
		var order order.Order
		// Get the order from the JSON body
		// Bind the JSON to the general Event struct
		if err := c.ShouldBindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		// Create an Event struct
		orderJSON, err := json.Marshal(order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal order"})
			return
		}
		event := common.NewEvent("OrderReceived", string(orderJSON))

		// Publish an event
		ctx := context.Background()
		producerErr := deps.Producer.Publish(ctx, event)
		if producerErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to marshal order: %v", producerErr),
			})
			return
		}
		log.Println("Published order event")

		c.JSON(http.StatusOK, gin.H{"status": "Order received", "eventId": event.EventId, "order": order})
	}
}
