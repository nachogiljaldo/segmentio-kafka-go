package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/nachogiljaldo/segmentio-kafka/pkg/producer"
	"github.com/nachogiljaldo/segmentio-kafka/pkg/tracing"
)

func main() {
	fmt.Println("Running producer")
	ctx := context.Background()
	close, _ := tracing.InitTraceExporter(ctx, "producer")
	defer close(ctx)
	producer, err := producer.NewProducer()
	if err != nil {
		panic(err)
	}

	for i := 0; i < 5; i++ {
		idx := i
		for j := 0; j < 5; j++ {
			msg := fmt.Sprintf("message by %d", idx)
			key, _ := uuid.NewUUID()
			if err := producer.Send(ctx, key.String(), msg); err != nil {
				fmt.Printf("Error: %+v\n")
				break
			}
			time.Sleep(1 * time.Second)
		}
	}

	defer producer.Close()
}
