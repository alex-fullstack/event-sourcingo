package kafka

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
)

type Writer struct {
	*kafka.Writer
}

func NewWriter(cfg *Config) *Writer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Address),
		Topic:    cfg.Topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &Writer{writer}
}

func (w *Writer) Publish(ctx context.Context, events []events.IntegrationEvent) error {
	messages := slices.Collect(func(yield func(kafka.Message) bool) {
		for _, event := range events {
			parsedId, _ := uuid.Parse(event.Id)
			key, _ := parsedId.MarshalBinary()
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
