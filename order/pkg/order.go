package order

import "time"

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
