package repositories

import (
	"context"
	"github.com/google/uuid"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/entities"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/subscriptions"
)

type TFACommitter interface {
	Begin(context.Context) (executor interface{}, err error)
	Commit(ctx context.Context, executor interface{}) error
	Rollback(ctx context.Context, executor interface{}) error
}

type EventStore interface {
	TFACommitter
	UpdateOrCreateAggregate(ctx context.Context, transactionId uuid.UUID, reader entities.AggregateReader, executor interface{}) (err error)
	GetAggregateEvents(ctx context.Context, id uuid.UUID, executor interface{}) ([]events.Event, error)
	GetNewEventsAndHistory(ctx context.Context, id uuid.UUID, firstSequenceId, lastSequenceId int64, executor interface{}) ([]events.Event, []events.Event, error)
	GetSubscription(ctx context.Context, executor interface{}) (*subscriptions.Subscription, error)
	UpdateSubscription(ctx context.Context, sub *subscriptions.Subscription, executor interface{}) error
}
