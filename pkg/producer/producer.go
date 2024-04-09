package producer

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	Send(key, value string) error
	Close() error
}

type producer struct {
	writer *kafka.Writer
	batch  *kafka.Batch
}

func NewProducer() (Producer, error) {
	w := &kafka.Writer{
		Addr:     kafka.TCP("localhost:29092"),
		Topic:    "test-topic",
		Balancer: &kafka.Hash{},
	}

	return producer{
		writer: w,
	}, nil
}

func (p producer) Close() error {
	if err := p.writer.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
		return err
	}
	return nil
}

func (p producer) Send(key, value string) error {
	err := p.writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(key),
			Value: []byte(value),
		},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
		return err
	}
	return nil
}
