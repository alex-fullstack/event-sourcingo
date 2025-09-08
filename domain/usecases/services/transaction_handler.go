package services

import (
	"context"
	"log/slog"

	"github.com/alex-fullstack/event-sourcingo/domain/entities"
	"github.com/alex-fullstack/event-sourcingo/domain/subscriptions"
	"github.com/alex-fullstack/event-sourcingo/domain/transactions"
	"github.com/alex-fullstack/event-sourcingo/domain/usecases/repositories"
	"github.com/google/uuid"
)

type TransactionHandler[T, S, P, K, E any] interface {
	Handle(
		ctx context.Context,
		transaction *transactions.Transaction,
		providerFn func(id uuid.UUID) entities.AggregateProvider[T, S, P, K],
	) error
}

type transactionHandler[T, S, P, K, E any] struct {
	eventStore   repositories.EventStore[T, S, E]
	eventHandler EventHandler[T, S, P, K]
	log          *slog.Logger
}

func NewTransactionHandler[T, S, P, K, E any](
	store repositories.EventStore[T, S, E],
	eventHandler EventHandler[T, S, P, K],
	log *slog.Logger,
) TransactionHandler[T, S, P, K, E] {
	return &transactionHandler[T, S, P, K, E]{
		eventStore:   store,
		eventHandler: eventHandler,
		log:          log,
	}
}

func (eh *transactionHandler[T, S, P, K, E]) Handle(
	ctx context.Context,
	transaction *transactions.Transaction,
	providerFn func(id uuid.UUID) entities.AggregateProvider[T, S, P, K],
) (err error) {
	commitExecutor, beginErr := eh.eventStore.Begin(ctx)
	if beginErr != nil {
		return beginErr
	}
	defer func() {
		if err != nil {
			commitErr := eh.eventStore.Rollback(ctx, commitExecutor)
			if commitErr != nil {
				err = commitErr
			}
		} else {
			err = eh.eventStore.Commit(ctx, commitExecutor)
		}
	}()
	sub, err := eh.eventStore.GetSubscription(ctx, commitExecutor)
	if err != nil {
		eh.log.ErrorContext(ctx, err.Error())
		return err
	}
	history, newEvents, err := eh.eventStore.GetNewEventsAndHistory(
		ctx,
		transaction.AggregateID,
		sub.LastSequenceID,
		transaction.SequenceID,
		commitExecutor,
	)
	if err != nil {
		eh.log.ErrorContext(ctx, err.Error())
		return err
	}
	provider := providerFn(transaction.AggregateID)
	err = provider.Build(history)
	if err != nil {
		eh.log.ErrorContext(ctx, err.Error())
		return err
	}
	err = eh.eventHandler.HandleEvents(ctx, provider, newEvents)
	if err != nil {
		eh.log.ErrorContext(ctx, err.Error())
		return err
	}

	return eh.eventStore.UpdateSubscription(
		ctx,
		&subscriptions.Subscription{LastSequenceID: transaction.SequenceID},
		commitExecutor,
	)
}
