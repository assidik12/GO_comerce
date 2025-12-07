package infrastructure

import (
	"log"

	"github.com/assidik12/go-restfull-api/config"
	"github.com/segmentio/kafka-go"
)

// NewKafkaWriter menginisialisasi Kafka Writer.
// Kita set AllowAutoTopicCreation ke true untuk development memudahkan.
func NewKafkaWriter(cfg config.Config) *kafka.Writer {
	conf := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaHost + ":" + cfg.KafkaPort),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}

	log.Println("connection to kafka success...")
	return conf
}
