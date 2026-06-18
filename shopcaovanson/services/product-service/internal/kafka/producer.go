package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
		},
	}
}

func (p *Producer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

type ProductCreatedEvent struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	SalePrice   *float64 `json:"sale_price,omitempty"`
	Category    string   `json:"category"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   string   `json:"created_at"`
}

type ProductUpdatedEvent struct {
	ID          string   `json:"id"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	SalePrice   *float64 `json:"sale_price,omitempty"`
	Category    *string  `json:"category,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
	UpdatedAt   string   `json:"updated_at"`
}

func (p *Producer) PublishProductCreated(ctx context.Context, event ProductCreatedEvent) error {
	return p.publish(ctx, "product.created", event.ID, event)
}

func (p *Producer) PublishProductUpdated(ctx context.Context, event ProductUpdatedEvent) error {
	return p.publish(ctx, "product.updated", event.ID, event)
}

func (p *Producer) publish(ctx context.Context, topic, key string, payload interface{}) error {
	if p == nil || p.writer == nil {
		return nil
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: body,
		Time:  time.Now().UTC(),
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka publish %s: %w", topic, err)
	}
	return nil
}
