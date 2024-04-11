package consumer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/nachogiljaldo/segmentio-kafka/pkg/tracing"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/nachogiljaldo/segmentio-kafka/pkg/consumer")

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
		MaxWait:        100 * time.Millisecond,
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
		m, err := p.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		carrier := tracing.NewMessageCarrier(&m)
		parentSpanContext := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

		// Create a span.
		attrs := []attribute.KeyValue{
			semconv.MessagingSystem("kafka"),
			semconv.MessagingDestinationKindTopic,
			semconv.MessagingDestinationName(m.Topic),
			semconv.MessagingOperationReceive,
			semconv.MessagingMessageID(strconv.FormatInt(m.Offset, 10)),
			semconv.MessagingKafkaSourcePartition(int(m.Partition)),
		}
		opts := []trace.SpanStartOption{
			trace.WithAttributes(attrs...),
			trace.WithSpanKind(trace.SpanKindConsumer),
		}
		newCtx, span := tracer.Start(parentSpanContext, fmt.Sprintf("%s receive", m.Topic), opts...)

		// Inject current span context, so consumers can use it to propagate span.
		otel.GetTextMapPropagator().Inject(newCtx, carrier)
		fmt.Printf("message at topic/partition/offset %v/%v/%v%v: %s = %s\n", m.Topic, m.Partition, m.Offset, carrier.Keys(), string(m.Key), string(m.Value))
		if err := p.reader.CommitMessages(ctx, m); err != nil {
			span.RecordError(err)
			log.Fatal("failed to commit messages:", err)
		}
		span.End()

	}
}
