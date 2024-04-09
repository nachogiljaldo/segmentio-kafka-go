package consumer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer interface {
	Subscribe() error
	Close() error
}

type consumer struct {
	reader *kafka.Reader
	batch  *kafka.Batch
}

func NewConsumer() (Consumer, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"localhost:29092"},
		GroupID:        "consumer-group-id",
		Topic:          "test-topic",
		MaxBytes:       10e6, // 10MB
		QueueCapacity:  100,
		CommitInterval: 1 * time.Second,
	})

	return consumer{
		reader: r,
	}, nil
}

func (p consumer) Close() error {
	if err := p.reader.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
		return err
	}
	return nil
}

func (p consumer) Subscribe() error {
	ctx := context.Background()
	for {
		m, err := p.reader.ReadMessage(context.Background())
		if err != nil {
			return err
		}
		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
		if err := p.reader.CommitMessages(ctx, m); err != nil {
			log.Fatal("failed to commit messages:", err)
		}
	}
}
