package services

import (
	"context"

	"github.com/alex-fullstack/event-sourcingo/domain/entities"
	"github.com/alex-fullstack/event-sourcingo/domain/events"
	"github.com/alex-fullstack/event-sourcingo/domain/usecases/repositories"
)

type EventHandler[T, S, P, K any] interface {
	HandleEvents(
		ctx context.Context,
		provider entities.AggregateProvider[T, S, P, K],
		events []events.Event[T],
	) error
}

type eventHandler[T, S, P, K any] struct {
	publisher repositories.Publisher[K]
}

func NewEventHandler[T, S, P, K any](publisher repositories.Publisher[K]) EventHandler[T, S, P, K] {
	return &eventHandler[T, S, P, K]{publisher: publisher}
}

func (eh *eventHandler[T, S, P, K]) HandleEvents(
	ctx context.Context,
	provider entities.AggregateProvider[T, S, P, K],
	newEvents []events.Event[T],
) error {
	integrationEvents := make([]events.IntegrationEvent[K], 0)
	for _, event := range newEvents {
		err := provider.ApplyChange(event)
		if err != nil {
			return err
		}
		integrationEvents = append(integrationEvents, provider.IntegrationEvent(event.Type))
	}
	return eh.publisher.Publish(ctx, integrationEvents)
}
