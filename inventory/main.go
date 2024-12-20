package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	events "github.com/tankcdr/ppe-kafka-go/events"
	kafka "github.com/tankcdr/ppe-kafka-go/kafka"

	"github.com/gin-gonic/gin"
)

// ProcessMessage processes the consumed Kafka message
func ProcessMessage(key, value []byte) error {
	log.Printf("Consumed message: Key=%s, Value=%s\n", string(key), string(value))

	var event *events.Event
	var order *events.Order
	var err error

	// Unmarshal the event
	if event, err = events.NewEventFromBytes(value); err != nil {
		log.Printf("Failed to unmarshal event: %v\n", err)
		return err
	}

	// Unmarshal the order
	if order, err = events.NewOrderFromBytes([]byte(event.EventBody)); err != nil {
		log.Printf("Failed to unmarshal order: %v\n", err)
		return err
	}

	// Enforce order idempotence
	// idea is that order ids are unique, but they are embedded in the order object

	return nil
}

func main() {
	// Define Kafka configuration
	kafkaConfig := kafka.KafkaConfig{
		Brokers: []string{"localhost:9092"}, // Replace with your Kafka broker(s)
		Topic:   "order-received",           // Replace with your topic
		GroupID: "inventory-group",          // Replace with your consumer group
	}

	// Create KafkaConsumer instance
	consumer := kafka.NewConsumer(kafkaConfig)
	defer consumer.Close()

	// Set up a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start REST server in a goroutine
	router := gin.Default()

	// /health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// /shutdown endpoint
	router.POST("/shutdown", func(c *gin.Context) {
		log.Println("Shutdown request received")
		cancel() // Signal the Kafka consumer to stop
		c.String(http.StatusOK, "Shutting down")
	})

	// Start the REST server
	go func() {
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("Failed to start REST server: %v\n", err)
		}
	}()

	// Handle graceful shutdown signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Start consuming Kafka messages
	go func() {
		log.Println("Starting Kafka consumer...")
		consumer.Consume(ctx, ProcessMessage)
	}()

	// Wait for the context to be canceled (e.g., via /shutdown or signal)
	<-ctx.Done()
	log.Println("Service is shutting down...")
}
