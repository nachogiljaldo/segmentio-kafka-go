package main

import (
	"context"
	"fmt"

	"github.com/nachogiljaldo/segmentio-kafka/pkg/consumer"
	"github.com/nachogiljaldo/segmentio-kafka/pkg/tracing"
)

func main() {
	fmt.Println("Running consumer")
	ctx := context.Background()
	close, _ := tracing.InitTraceExporter(ctx, "consumer")
	defer close(ctx)
	consumer, err := consumer.NewConsumer()
	if err != nil {
		panic(err)
	}

	if err := consumer.Subscribe(); err != nil {
		panic(err)
	}

	defer consumer.Close()
}
