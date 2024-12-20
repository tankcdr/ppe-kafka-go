package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

type EventType int

const (
	OrderReceived EventType = iota
	OrderConfirmed
)

var orderStatus = map[EventType]string{
	OrderReceived:  "OrderReceived",
	OrderConfirmed: "OrderConfirmed",
}

type Event struct {
	EventId   string `json:"eventId"`
	EventName string `json:"eventName"`
	Timestamp string `json:"timestamp"`
	EventBody string `json:"eventBody"`
}

func NewEvent(eventType EventType, eventBody string) *Event {
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
		EventName: orderStatus[eventType],
		Timestamp: timestamp,
		EventBody: eventBody,
	}
}

func NewEventFromBytes(value []byte) (*Event, error) {
	event := &Event{}
	if err := json.Unmarshal(value, event); err != nil {
		return nil, err
	}
	return event, nil
}

/****************************************************************************************
 * Order implementation
 * This is expected input into the system and will be used to process orders
 ****************************************************************************************/
// OrderItem represents an item in the order.
type OrderItem struct {
	ItemID   string  `json:"itemId"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// OrderBody represents the body of the OrderReceived event.
type Order struct {
	OrderID     string      `json:"orderId"`
	CustomerID  string      `json:"customerId"`
	OrderDate   time.Time   `json:"orderDate"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"totalAmount"`
}

func NewOrderFromBytes(value []byte) (*Order, error) {
	order := &Order{}
	if err := json.Unmarshal(value, order); err != nil {
		return nil, err
	}
	return order, nil
}
