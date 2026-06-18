package kafka

// Order-service only produces Kafka events.

type Consumer struct{}

func NewConsumer(_ []string, _ string, _ string) *Consumer {
	return &Consumer{}
}

func (c *Consumer) Close() error {
	return nil
}
