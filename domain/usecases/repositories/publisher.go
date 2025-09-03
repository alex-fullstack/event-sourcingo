package repositories

import (
	"context"

	"github.com/alex-fullstack/event-sourcingo/domain/events"
)

type Publisher interface {
	Publish(context.Context, []events.IntegrationEvent) error
}
