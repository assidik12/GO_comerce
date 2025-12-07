package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

// Producer adalah abstract contract.
// Service cuma butuh tau cara Publish, ga perlu tau detail Kafka/RabbitMQ.
type Producer interface {
	Publish(ctx context.Context, topic string, message interface{}) error
}

type KafkaProducer struct {
	Writer *kafka.Writer
}

// NewKafkaProducer constructor untuk Wire
func NewKafkaProducer(writer *kafka.Writer) *KafkaProducer {
	return &KafkaProducer{
		Writer: writer,
	}
}

// Publish mengimplementasikan domainEvent.Producer
func (k *KafkaProducer) Publish(ctx context.Context, topic string, message interface{}) error {
	// 1. Serialize message ke JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 2. Kirim ke Kafka
	err = k.Writer.WriteMessages(ctx,
		kafka.Message{
			Topic: topic,
			Value: payload,
		},
	)

	if err != nil {
		log.Printf("[Kafka] Failed to publish to topic %s: %v", topic, err)
		return err
	}

	log.Printf("[Kafka] Successfully published to topic %s", topic)
	return nil
}

// Pastikan KafkaProducer mengimplementasikan interface domainEvent.Producer
var _ Producer = (*KafkaProducer)(nil)
