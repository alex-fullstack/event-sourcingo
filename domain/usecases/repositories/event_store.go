package repositories

import (
	"context"

	"github.com/alex-fullstack/event-sourcingo/domain/entities"
	"github.com/alex-fullstack/event-sourcingo/domain/events"
	"github.com/alex-fullstack/event-sourcingo/domain/subscriptions"
	"github.com/google/uuid"
)

type TFACommitter[E any] interface {
	Begin(context.Context) (executor E, err error)
	Commit(ctx context.Context, executor E) error
	Rollback(ctx context.Context, executor E) error
}

type EventStore[T, S, E any] interface {
	TFACommitter[E]
	UpdateOrCreateAggregate(
		ctx context.Context,
		transactionID uuid.UUID,
		reader entities.AggregateReader[T],
		snapshot S,
		executor E,
	) (snapshotCount int, err error)
	GetLastSnapshot(
		ctx context.Context,
		id uuid.UUID,
		executor E,
	) (int, S, error)
	GetHistory(
		ctx context.Context,
		id uuid.UUID,
		fromVersion int,
		executor E,
	) ([]events.Event[T], error)
	GetNewEventsAndHistory(
		ctx context.Context,
		id uuid.UUID,
		firstSequenceID, lastSequenceID int64,
		executor E,
	) ([]events.Event[T], []events.Event[T], error)
	GetSubscription(ctx context.Context, executor E) (*subscriptions.Subscription, error)
	UpdateSubscription(
		ctx context.Context,
		sub *subscriptions.Subscription,
		executor E,
	) error
}
