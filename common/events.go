package common

type Event struct {
	EventId   string `json:"eventId"`
	EventName string `json:"eventName"`
	Timestamp string `json:"timestamp"`
	EventBody string `json:"eventBody"`
}
