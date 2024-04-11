package producer

import (
	"context"
	"fmt"
	"log"

	"github.com/nachogiljaldo/segmentio-kafka/pkg/tracing"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type Producer interface {
	Send(ctx context.Context, key, value string) error
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

var tracer = otel.Tracer("github.com/nachogiljaldo/segmentio-kafka/pkg/producer")

func (p producer) Close() error {
	if err := p.writer.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
		return err
	}
	return nil
}

func (p producer) Send(ctx context.Context, key, value string) error {
	msg := kafka.Message{
		Key:   []byte(key),
		Value: []byte(value),
	}
	carrier := tracing.NewMessageCarrier(&msg)
	ctx = otel.GetTextMapPropagator().Extract(context.Background(), carrier)

	// Create a span.
	attrs := []attribute.KeyValue{
		semconv.MessagingSystem("kafka"),
		semconv.MessagingDestinationKindTopic,
		semconv.MessagingDestinationName(msg.Topic),
		semconv.MessagingMessagePayloadSizeBytes(len(msg.Value)),
		semconv.MessagingOperationPublish,
	}
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindProducer),
	}
	ctx, span := tracer.Start(ctx, fmt.Sprintf("%s publish", msg.Topic), opts...)

	// Inject current span context, so consumers can use it to propagate span.
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	defer span.End()

	err := p.writer.WriteMessages(context.Background(),
		msg,
	)
	if err != nil {
		span.RecordError(err)
		log.Fatal("failed to write messages:", err)
		return err
	}
	return nil
}
