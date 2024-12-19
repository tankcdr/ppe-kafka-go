package common

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

type Event struct {
	EventId   string `json:"eventId"`
	EventName string `json:"eventName"`
	Timestamp string `json:"timestamp"`
	EventBody string `json:"eventBody"`
}

func NewEvent(eventName string, eventBody string) *Event {
	// Generate a new UUID for eventId
	u, err := uuid.NewV4()
	if err != nil {
		fmt.Printf("Failed to generate UUID: %v\n", err)
		return nil
	}

	// Get current timestamp
	now := time.Now()
	timestamp := now.Format(time.RFC3339)

	return &Event{
		EventId:   u.String(),
		EventName: eventName,
		Timestamp: timestamp,
		EventBody: eventBody,
	}
}
