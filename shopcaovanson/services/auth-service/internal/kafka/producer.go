package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/IBM/sarama"
)

type UserRegisteredEvent struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type VerificationRequestedEvent struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Token     string `json:"token"`
	VerifyURL string `json:"verify_url"`
}

type Producer struct {
	producer sarama.SyncProducer
	registeredTopic string
	verifyTopic     string
}

func NewProducer(brokers string) (*Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForLocal

	brokerList := strings.Split(brokers, ",")
	for i := range brokerList {
		brokerList[i] = strings.TrimSpace(brokerList[i])
	}

	producer, err := sarama.NewSyncProducer(brokerList, cfg)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	return &Producer{
		producer:        producer,
		registeredTopic: "user.registered",
		verifyTopic:     "user.verification_requested",
	}, nil
}

func (p *Producer) publish(topic, key string, payload []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(payload),
	}
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	log.Printf("kafka: published %s key=%s partition=%d offset=%d", topic, key, partition, offset)
	return nil
}

func (p *Producer) PublishUserRegistered(event UserRegisteredEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return p.publish(p.registeredTopic, event.UserID, payload)
}

func (p *Producer) PublishVerificationRequested(event VerificationRequestedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return p.publish(p.verifyTopic, event.UserID, payload)
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
