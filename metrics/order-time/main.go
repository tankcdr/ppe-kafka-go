package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	db "github.com/tankcdr/ppe-kafka-go/db"
	events "github.com/tankcdr/ppe-kafka-go/events"
	kafka "github.com/tankcdr/ppe-kafka-go/kafka"
)

var (
	// timeToShip measures the distribution of durations (in seconds)
	// between the "Order Received" event and the "Order Picked & Packed" event.
	timeToShip = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "order_time_to_ship_seconds",
		Help:    "Time (in seconds) between order received and order picked & packed",
		Buckets: prometheus.DefBuckets, // may customize this if needed
	})

	// store keeps track of orders and when they were first received.
	// key: order_id, value: time when the order was received
	store = db.NewSimpleInMemoryDatabase[string, time.Time]()
)

func init() {
	// Register our histogram with the default registry
	prometheus.MustRegister(timeToShip)
}

// Config holds the environment configuration
type Config struct {
	Broker string `env:"KAFKA_BROKER" envDefault:"localhost:29092"`
	//consuming order-received and order-picked-packed topics
	OrderReceivedTopic     string `env:"KAFKA_ORDER_RECEIVED" envDefault:"order-received"`
	OrderPickedPackedTopic string `env:"KAFKA_ORDER_PICKED_PACKED" envDefault:"order-picked-packed"`
}

// AppDependencies holds shared dependencies like Kafka producers
type AppDependencies struct {
	Config Config
}

func main() {
	// Set up a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Load configuration
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Start the Kafka consumers in separate goroutines
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumeOrderReceived(&cfg, ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		consumeOrderPickedPacked(&cfg, ctx)
	}()

	// Start a Gin HTTP server to expose /metrics
	router := gin.Default()

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	// /shutdown endpoint
	router.POST("/shutdown", func(c *gin.Context) {
		log.Println("Shutdown request received")
		cancel() // Signal the Kafka consumer to stop
		c.JSON(200, gin.H{"status": "Shutting down"})
	})

	// Start the REST server
	wg.Add(1)
	go func() {
		defer wg.Done()
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

	// Wait for the context to be canceled (e.g., via /shutdown or signal)
	<-ctx.Done()
	log.Println("Service is shutting down...")

	// Wait for all goroutines to finish
	wg.Wait()
	log.Println("All goroutines have exited. Service has shut down.")
}

func consumeOrderReceived(config *Config, context context.Context) {

	orderConsumerConfig := kafka.KafkaConfig{
		Brokers: []string{config.Broker},
		Topic:   config.OrderReceivedTopic,
		GroupID: "metrics-group",
	}

	// Create KafkaConsumer instance
	consumer := kafka.NewConsumer(orderConsumerConfig)
	defer consumer.Close()

	log.Println("Listening for Order Received events...")
	consumer.Consume(context, func(key, value []byte) error {
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
		eventName := events.OrderStatus[events.OrderReceived]
		if event.EventName != eventName {
			log.Printf("Not of type %s. Instead event type is %s.\n", eventName, event.EventName)
			return nil
		}

		//convert event time to time.Time
		eventTime, err := toTime(event.Timestamp)

		if err != nil {
			log.Printf("Failed to convert event time to time.Time: %v\n", err)
			return err
		}
		//store in db
		store.Add(order.OrderID, eventTime)
		log.Printf("Stored in database: Order %s received at %s\n", order.OrderID, eventTime)

		return nil
	})
}

func consumeOrderPickedPacked(config *Config, context context.Context) {

	orderConsumerConfig := kafka.KafkaConfig{
		Brokers: []string{config.Broker},
		Topic:   config.OrderPickedPackedTopic,
		GroupID: "metrics-group",
	}

	// Create KafkaConsumer instance
	consumer := kafka.NewConsumer(orderConsumerConfig)
	defer consumer.Close()

	fmt.Println("Listening for Order Picked & Packed events...")
	consumer.Consume(context, func(key, value []byte) error {
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
		eventName := events.OrderStatus[events.OrderPickedPacked]
		if event.EventName != eventName {
			log.Printf("Not of type %s. Instead event type is %s.\n", eventName, event.EventName)
			return nil
		}

		//convert event time to time.Time
		eventTime, err := toTime(event.Timestamp)

		if err != nil {
			log.Printf("Failed to convert event time to time.Time: %v\n", err)
			return err
		}

		// Calculate duration if we have an entry in the store
		startTime, ok := store.Get(order.OrderID)

		if ok {
			duration := eventTime.Sub(startTime).Seconds()
			timeToShip.Observe(duration) // Record in histogram
			fmt.Printf("Order %s completed in %.2f seconds\n", order.OrderID, duration)
			store.Delete(order.OrderID) // Clean up memory
		}

		return nil
	})
}

func toTime(input string) (time.Time, error) {
	// Parse the time in RFC3339 format
	t, err := time.Parse(time.RFC3339, input)

	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time: %v", err)
	}

	return t, nil
}
