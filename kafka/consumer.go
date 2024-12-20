package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a new KafkaConsumer instance.
func NewConsumer(config KafkaConfig) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     config.Brokers,
			Topic:       config.Topic,
			GroupID:     config.GroupID,
			StartOffset: kafka.FirstOffset, // Change to kafka.LastOffset if needed
		}),
	}
}

// Consume starts consuming messages and calls the handler for each message.
func (c *KafkaConsumer) Consume(ctx context.Context, handler func(key, value []byte) error) {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v\n", err)
			continue
		}

		if err := handler(msg.Key, msg.Value); err != nil {
			log.Printf("Error handling message: %v\n", err)
		}
	}
}

// Close closes the Kafka consumer.
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
