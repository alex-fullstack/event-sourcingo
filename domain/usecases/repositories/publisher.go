package repositories

import (
	"context"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
)

type Publisher interface {
	Publish(context.Context, []events.IntegrationEvent) error
}
