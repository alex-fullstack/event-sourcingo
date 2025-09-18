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
	) error
	GetSnapshot(
		ctx context.Context,
		id uuid.UUID,
		versionAfter *int,
		executor E,
	) (int, S, error)
	GetEvents(
		ctx context.Context,
		id uuid.UUID,
		fromVersion int,
		toVersion *int,
		executor E,
	) ([]events.Event[T], error)
	GetUnhandledEvents(
		ctx context.Context,
		id uuid.UUID,
		firstSequenceID, lastSequenceID int64,
		executor E,
	) ([]events.Event[T], error)
	GetSubscription(ctx context.Context, executor E) (*subscriptions.Subscription, error)
	UpdateSubscription(
		ctx context.Context,
		sub *subscriptions.Subscription,
		executor E,
	) error
}
