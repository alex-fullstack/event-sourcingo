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

type TransactionHandler interface {
	Handle(
		ctx context.Context,
		transaction *transactions.Transaction,
		providerFn func(id uuid.UUID) entities.AggregateProvider,
	) error
}

type transactionHandler struct {
	eventStore   repositories.EventStore
	eventHandler EventHandler
	log          *slog.Logger
}

func NewTransactionHandler(
	store repositories.EventStore,
	eventHandler EventHandler,
	log *slog.Logger,
) TransactionHandler {
	return &transactionHandler{eventStore: store, eventHandler: eventHandler, log: log}
}

func (eh *transactionHandler) Handle(
	ctx context.Context,
	transaction *transactions.Transaction,
	providerFn func(id uuid.UUID) entities.AggregateProvider,
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
	err = eh.eventHandler.HandleEvents(ctx, history, newEvents, providerFn(transaction.AggregateID))
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
