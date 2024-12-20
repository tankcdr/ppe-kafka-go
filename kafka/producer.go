package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/tankcdr/ppe-kafka-go/events"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

// NewProducer creates a new KafkaProducer instance.
func NewProducer(config KafkaConfig) *KafkaProducer {
	return &KafkaProducer{
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers: config.Brokers,
			Topic:   config.Topic,
		}),
	}
}

// Publish sends a message to the Kafka topic.
func (p *KafkaProducer) Publish(ctx context.Context, event *events.Event) error {
	// Serialize the event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v\n", err)
		return err
	}

	// Create a Kafka message
	msg := kafka.Message{
		Key:   []byte(event.EventId), // Use EventId as the key, or maybe the order id?
		Value: eventJSON,
	}

	// Publish the message to Kafka
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Printf("Failed to publish message: %v\n", err)
		return err
	}

	log.Printf("Published event to Kafka: %s\n", eventJSON)
	return nil
}

// Close closes the Kafka producer.
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
