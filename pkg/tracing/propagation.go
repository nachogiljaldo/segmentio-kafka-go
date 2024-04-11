package tracing

import (
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/propagation"
)

type messageCarrier struct {
	message *kafka.Message
}

func NewMessageCarrier(message *kafka.Message) propagation.TextMapCarrier {
	return messageCarrier{
		message: message,
	}
}

func (m messageCarrier) Get(key string) string {
	for _, header := range m.message.Headers {
		if header.Key == key {
			return string(header.Value)
		}
	}
	return ""
}

// Set stores the key-value pair.
func (m messageCarrier) Set(key string, value string) {
	for _, header := range m.message.Headers {
		if header.Key == key {
			header.Value = []byte(value)
			return
		}
	}
	m.message.Headers = append(m.message.Headers, kafka.Header{
		Key:   key,
		Value: []byte(value),
	})
}

// Keys lists the keys stored in this carrier.
func (m messageCarrier) Keys() []string {
	keys := make([]string, 0, len(m.message.Headers))
	for _, header := range m.message.Headers {
		keys = append(keys, header.Key)
	}
	return keys
}
