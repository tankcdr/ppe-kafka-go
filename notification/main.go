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
	events "github.com/tankcdr/ppe-kafka-go/events"
	kafka "github.com/tankcdr/ppe-kafka-go/kafka"
)

// Config holds the environment configuration
type Config struct {
	Broker                 string `env:"KAFKA_BROKER" envDefault:"localhost:29092"`
	OrderConfirmedTopic    string `env:"KAFKA_ORDER_CONFIRMED" envDefault:"order-confirmed"`
	OrderNotificationTopic string `env:"KAFKA_ORDER_NOTFIFICATION" envDefault:"order-notification"`
	ErrorTopic             string `env:"KAFKA_ERROR" envDefault:"order-error"`
}

// AppDependencies holds shared dependencies like Kafka producers
type AppDependencies struct {
	Producer *kafka.KafkaProducer
	Config   Config
}

type KafkaProducers struct {
	NotificationProducer *kafka.KafkaProducer
	ErrorProducer        *kafka.KafkaProducer
}

// ProcessMessage processes the consumed Kafka message
func ProcessMessageWrapper(db *db.SimpleDatabase, producers *KafkaProducers) func(key, value []byte) error {
	return func(key, value []byte) error {
		log.Printf("Consumed message: Key=%s, Value=%s\n", string(key), string(value))

		var event *events.Event
		var notification *events.Notification
		var err error
		context := context.Background()

		// Unmarshal the event
		if event, err = events.NewEventFromBytes(value); err != nil {
			log.Printf("Failed to unmarshal event: %v\n", err)
			return err
		}
		// Check if the event is an Notification event
		if event.EventName != events.OrderStatus[events.OrderConfirmed] {
			log.Printf("Not of type Notification. Instead event type is %s.\n", event.EventName)
			return nil
		}

		// Unmarshal the Notification
		if notification, err = events.NewNotificationFromBytes([]byte(event.EventBody)); err != nil {
			log.Printf("Failed to unmarshal notification: %v\n", err)
			return err
		}

		// create a unqiue key for the notification using order id and type
		uniqueKey := events.NotificationStatus[events.NotificationType(notification.Type)] + ":" + notification.OrderID

		// Enforce order idempotence
		// idea is that order ids are unique, but they are embedded in the order object
		// using an in memory store, but would want a real db for this
		if db.Exists(uniqueKey) {
			log.Printf("Notification %s is a duplicate\n", uniqueKey)
			errorString := fmt.Sprintf("Notification %s is a duplicate", uniqueKey)
			event.ErrorMessage = &errorString

			// Publish an Error event to Kafka
			if err := producers.ErrorProducer.Publish(context, event); err != nil {
				log.Printf("Failed to produce Error event: %v\n", err)
				return fmt.Errorf("Failed to produce Error event: %v", err)
			}

			return fmt.Errorf(errorString)
		}
		db.Add(uniqueKey)
		log.Printf("Notification %s is unique\n", uniqueKey)

		log.Printf("Sucessfully processed notification event: %v\n", notification)

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
		NotificationProducer: kafka.NewProducer(kafka.KafkaConfig{
			Brokers: []string{cfg.Broker},
			Topic:   cfg.OrderNotificationTopic,
		}),
		ErrorProducer: kafka.NewProducer(kafka.KafkaConfig{
			Brokers: []string{cfg.Broker},
			Topic:   cfg.ErrorTopic,
		}),
	}
	defer producers.NotificationProducer.Close()
	defer producers.ErrorProducer.Close()

	// Define Kafka configuration
	kafkaConfigConsumer := kafka.KafkaConfig{
		Brokers: []string{cfg.Broker},
		Topic:   cfg.OrderConfirmedTopic,
		GroupID: "notification-group",
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
