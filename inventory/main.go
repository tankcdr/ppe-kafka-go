package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"

	db "github.com/tankcdr/ppe-kafka-go/db"
	errors "github.com/tankcdr/ppe-kafka-go/error"
	events "github.com/tankcdr/ppe-kafka-go/events"
	kafka "github.com/tankcdr/ppe-kafka-go/kafka"
)

// Config holds the environment configuration
type Config struct {
	Broker              string `env:"KAFKA_BROKER" envDefault:"localhost:29092"`
	OrderReceivedTopic  string `env:"KAFKA_ORDER_RECEIVED" envDefault:"order-received"`
	OrderConfirmedTopic string `env:"KAFKA_ORDER_CONFIRMED" envDefault:"order-confirmed"`
	ErrorTopic          string `env:"KAFKA_ERROR" envDefault:"error"`
}

// AppDependencies holds shared dependencies like Kafka producers
type AppDependencies struct {
	Producer *kafka.KafkaProducer
	Config   Config
}

type KafkaProducers struct {
	OrderConfirmedProducer *kafka.KafkaProducer
	ErrorProducer          *kafka.KafkaProducer
}

// ProcessMessage processes the consumed Kafka message
func ProcessMessageWrapper(db *db.SimpleDatabase, producers *KafkaProducers) func(key, value []byte) error {
	return func(key, value []byte) error {
		log.Printf("Consumed message: Key=%s, Value=%s\n", string(key), string(value))

		var event *events.Event
		var order *events.Order
		var err error
		context := context.Background()

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
		// using an in memory store, but would want a real db for this
		if db.Exists(order.OrderID) {
			errorString := fmt.Sprintf("Order %s is a duplicate", order.OrderID)
			return errors.HandleError(context, event, errorString)
		}
		db.Add(order.OrderID)
		log.Printf("Order %s is unique\n", order.OrderID)

		// Publish a new OrderConfirmed event to Kafka
		confirmedEvent := events.NewEvent(events.OrderConfirmed, event.EventBody)
		if err := producers.OrderConfirmedProducer.Publish(context, confirmedEvent); err != nil {
			errorString := fmt.Sprintf("Failed to produce OrderConfirmed event: %v\n", err)
			return errors.HandleError(context, event, errorString)
		}
		log.Printf("Published OrderConfirmed event: %v\n", confirmedEvent)

		return nil
	}
}

func main() {
	// Load configuration
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create "database" for in-memory idempotence check
	db := db.NewSimpleDatabase()

	// Create Kafka producers
	producers := KafkaProducers{
		OrderConfirmedProducer: kafka.NewProducer(kafka.KafkaConfig{
			Brokers: []string{cfg.Broker},
			Topic:   cfg.OrderConfirmedTopic,
		}),
		ErrorProducer: kafka.NewProducer(kafka.KafkaConfig{
			Brokers: []string{cfg.Broker},
			Topic:   cfg.ErrorTopic,
		}),
	}
	defer producers.OrderConfirmedProducer.Close()
	defer producers.ErrorProducer.Close()

	// Define Kafka configuration
	kafkaConfigConsumer := kafka.KafkaConfig{
		Brokers: []string{cfg.Broker},
		Topic:   cfg.OrderReceivedTopic,
		GroupID: "inventory-group",
	}

	// Create KafkaConsumer instance
	consumer := kafka.NewConsumer(kafkaConfigConsumer)
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
		consumer.Consume(ctx, ProcessMessageWrapper(db, &producers))
	}()

	// Wait for the context to be canceled (e.g., via /shutdown or signal)
	<-ctx.Done()
	log.Println("Service is shutting down...")
}
