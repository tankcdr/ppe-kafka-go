package kafka

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string // Used for consumers
}
