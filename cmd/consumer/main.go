package main

import (
	"fmt"

	"github.com/nachogiljaldo/segmentio-kafka/pkg/consumer"
)

func main() {
	fmt.Printf("Running consumer")
	consumer, err := consumer.NewConsumer()
	if err != nil {
		panic(err)
	}

	if err := consumer.Subscribe(); err != nil {
		panic(err)
	}

	defer consumer.Close()
}
