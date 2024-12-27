package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	db "github.com/tankcdr/ppe-kafka-go/db"
	events "github.com/tankcdr/ppe-kafka-go/events"
	kafka "github.com/tankcdr/ppe-kafka-go/kafka"
)

var (
	// errorCounter tracks total error messages consumed
	errorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_error_total",
			Help: "Total number of error events consumed from Kafka",
		},
		[]string{"order"},
	)
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(errorCounter)
}

// Config holds the environment configuration
type Config struct {
	Broker string `env:"KAFKA_BROKER" envDefault:"localhost:29092"`
	//consuming from the error topic
	ErrorTopic string `env:"KAFKA_ERROR" envDefault:"order-error"`
}

// AppDependencies holds shared dependencies like Kafka producers
type AppDependencies struct {
	Config Config
}

// process the metric
func ProcessMetricWrapper(db *db.SimpleDatabase) func(key, value []byte) error {
	return func(key, value []byte) error {
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

		// Check if the event is an OrderConfirmed event
		// Inventory service publishes the OrderConfirmed event
		if event.EventName != events.OrderStatus[events.Error] {
			log.Printf("Not of type %s. Instead event type is %s.\n", events.OrderStatus[events.Error], event.EventName)
			return nil
		}

		// create a unqiue key for the notification using order id and type
		uniqueKey := event.EventId
		// Enforce order idempotence
		// idea is that order ids are unique, but they are embedded in the order object
		// using an in memory store, but would want a real db for this
		if db.Exists(uniqueKey) {
			log.Printf("Error Event Id %s is a duplicate", uniqueKey)
			return nil
		}
		db.Add(uniqueKey)
		log.Printf("Error Event Id %s is unique\n", uniqueKey)

		// Increment the Prometheus counter for each message
		errorCounter.With(prometheus.Labels{"order": order.OrderID}).Inc()
		fmt.Printf("Updated error count for order %s\n", order.OrderID)

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

	// Define Kafka configuration
	kafkaConfigConsumer := kafka.KafkaConfig{
		Brokers: []string{cfg.Broker},
		Topic:   cfg.ErrorTopic,
		GroupID: "metrics-group",
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
		c.JSON(200, gin.H{"status": "OK"})
	})

	// prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// /shutdown endpoint
	router.POST("/shutdown", func(c *gin.Context) {
		log.Println("Shutdown request received")
		cancel() // Signal the Kafka consumer to stop
		c.JSON(200, gin.H{"status": "Shutting down"})
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
		consumer.Consume(ctx, ProcessMetricWrapper(db))
	}()

	// Wait for the context to be canceled (e.g., via /shutdown or signal)
	<-ctx.Done()
	log.Println("Service is shutting down...")
}
