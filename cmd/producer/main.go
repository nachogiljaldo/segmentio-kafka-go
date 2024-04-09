package main

import (
	"fmt"

	"github.com/nachogiljaldo/segmentio-kafka/pkg/producer"
)

func main() {
	fmt.Printf("Running producer")
	producer, err := producer.NewProducer()
	if err != nil {
		panic(err)
	}

	for {
		if err := producer.Send("foo", "bar"); err != nil {
			panic(err)
		}
	}

	defer producer.Close()
}
