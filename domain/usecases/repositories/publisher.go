package repositories

import (
	"context"

	"github.com/alex-fullstack/event-sourcingo/domain/events"
)

type Publisher[T any] interface {
	Publish(context.Context, []events.IntegrationEvent[T]) error
}
