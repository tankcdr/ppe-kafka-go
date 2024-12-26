package error

import (
	"context"
	"fmt"
	"log"

	"github.com/tankcdr/ppe-kafka-go/events"
	"github.com/tankcdr/ppe-kafka-go/kafka"
)

// HandleError processes an error by logging it, updating the event, publishing the error to Kafka, and returning the formatted error.
func HandleError(ctx context.Context, event *events.Event, producer *kafka.KafkaProducer, customMessage string) error {
	// Log the error
	log.Printf("%s\n", customMessage)

	err := fmt.Errorf(customMessage)

	// Update the event with the error message
	errorString := fmt.Sprintf("%s: %v", customMessage, err)
	event.ErrorMessage = &errorString

	// Publish an error event to Kafka
	if publishErr := producer.Publish(ctx, event); publishErr != nil {
		log.Printf("Failed to produce Error event: %v\n", publishErr)
		return fmt.Errorf("Failed to produce Error event: %v", publishErr)
	}

	// Return the formatted error
	return fmt.Errorf(errorString)
}
