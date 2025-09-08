package kafka

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/alex-fullstack/event-sourcingo/domain/events"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Writer[T any] struct {
	*kafka.Writer
}

func NewWriter[T any](cfg *Config) *Writer[T] {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Address),
		Topic:    cfg.Topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &Writer[T]{writer}
}

func (w *Writer[T]) Publish(ctx context.Context, events []events.IntegrationEvent[T]) error {
	messages := slices.Collect(func(yield func(kafka.Message) bool) {
		for _, event := range events {
			parsedID, _ := uuid.Parse(event.ID)
			key, _ := parsedID.MarshalBinary()
			value, _ := json.Marshal(event)
			message := kafka.Message{
				Key:   key,
				Value: value,
			}
			if !yield(message) {
				return
			}
		}
	})
	return w.WriteMessages(ctx, messages...)
}
