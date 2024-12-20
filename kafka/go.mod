module github.com/tankcdr/ppe-kafka-go/kafka

go 1.23.2

require (
	github.com/segmentio/kafka-go v0.4.47
	github.com/tankcdr/ppe-kafka-go/events v0.0.0
)

require (
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
)

replace github.com/tankcdr/ppe-kafka-go/events => ../events
