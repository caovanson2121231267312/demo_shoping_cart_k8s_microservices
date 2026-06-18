package kafka

// Product-service only produces Kafka events; no consumers are required.
// Consumer wiring is handled by search-service and notification-service.

type Consumer struct{}

func NewConsumer(_ []string, _ string, _ string) *Consumer {
	return &Consumer{}
}

func (c *Consumer) Close() error {
	return nil
}
