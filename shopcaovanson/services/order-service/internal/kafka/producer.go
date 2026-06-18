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
		},
	}
}

func (p *Producer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

type OrderCreatedItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type OrderCreatedEvent struct {
	OrderID         string             `json:"order_id"`
	OrderNumber     string             `json:"order_number"`
	UserID          string             `json:"user_id"`
	UserEmail       string             `json:"user_email"`
	Items           []OrderCreatedItem `json:"items"`
	SubtotalAmount  float64            `json:"subtotal_amount"`
	DiscountAmount  float64            `json:"discount_amount"`
	CouponCode      string             `json:"coupon_code"`
	TotalAmount     float64            `json:"total_amount"`
	ShippingName    string             `json:"shipping_name"`
	ShippingPhone   string             `json:"shipping_phone"`
	ShippingAddress string             `json:"shipping_address"`
	CreatedAt       string             `json:"created_at"`
}

func (p *Producer) PublishOrderCreated(ctx context.Context, event OrderCreatedEvent) error {
	return p.publish(ctx, "order.created", event.OrderID, event)
}

func (p *Producer) PublishInvoiceGenerate(ctx context.Context, event OrderCreatedEvent) error {
	return p.publish(ctx, "order.invoice.generate", event.OrderID, event)
}

func (p *Producer) publish(ctx context.Context, topic, key string, event OrderCreatedEvent) error {
	if p == nil || p.writer == nil {
		return nil
	}
	body, err := json.Marshal(event)
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
