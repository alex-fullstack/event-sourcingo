package services

import (
	"context"

	"github.com/alex-fullstack/event-sourcingo/domain/entities"
	"github.com/alex-fullstack/event-sourcingo/domain/events"
	"github.com/alex-fullstack/event-sourcingo/domain/usecases/repositories"
)

type EventHandler interface {
	HandleEvents(ctx context.Context, history, newEvents []events.Event, provider entities.AggregateProvider) error
}

type eventHandler struct {
	publisher repositories.Publisher
}

func NewEventHandler(publisher repositories.Publisher) EventHandler {
	return &eventHandler{publisher: publisher}
}

func (eh *eventHandler) HandleEvents(ctx context.Context, history, newEvents []events.Event, aggregate entities.AggregateProvider) error {
	err := aggregate.Build(history)
	if err != nil {
		return err
	}
	integrationEvents := make([]events.IntegrationEvent, 0)
	for _, event := range newEvents {
		err := aggregate.ApplyChanges([]events.Event{event})
		if err != nil {
			return err
		}
		integrationEvent, err := aggregate.IntegrationEvent(event.Type)
		if err != nil {
			return err
		}
		integrationEvents = append(integrationEvents, integrationEvent)
	}
	return eh.publisher.Publish(ctx, integrationEvents)
}
