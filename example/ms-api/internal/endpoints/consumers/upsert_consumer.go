package consumers

import (
	"api/internal/infrastructure/kafka"
	"context"
	"encoding/json"

	"github.com/google/uuid"
	kafkaGo "github.com/segmentio/kafka-go"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/endpoints"
)

type upsertConsumer struct {
	*endpoints.Endpoint
}

func newUpsertConsumer(
	baseContext context.Context,
	cfg kafkaGo.ReaderConfig,
	handler func(ctx context.Context, message kafkaGo.Message) error,
) endpoints.EndpointStarter {
	reader := kafka.NewReader(
		cfg,
		handler,
		func() context.Context {
			return baseContext
		},
	)
	return &upsertConsumer{
		Endpoint: endpoints.NewEndpoint(
			reader.StartListen,
			reader.Shutdown,
		),
	}
}

func convert(message kafkaGo.Message) (uuid.UUID, events.IntegrationEvent, error) {
	id, err := uuid.FromBytes(message.Key)
	if err != nil {
		return [16]byte{}, events.IntegrationEvent{}, err
	}
	var ev events.IntegrationEvent
	err = json.Unmarshal(message.Value, &ev)
	if err != nil {
		return [16]byte{}, events.IntegrationEvent{}, err
	}
	return id, ev, err
}
