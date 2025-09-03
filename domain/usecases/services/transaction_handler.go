package services

import (
	"context"
	"log"

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
}

func NewTransactionHandler(store repositories.EventStore, eventHandler EventHandler) TransactionHandler {
	return &transactionHandler{eventStore: store, eventHandler: eventHandler}
}

func (eh *transactionHandler) Handle(ctx context.Context, transaction *transactions.Transaction, providerFn func(id uuid.UUID) entities.AggregateProvider) (err error) {
	commitExecutor, beginErr := eh.eventStore.Begin(ctx)
	if beginErr != nil {
		err = beginErr
		return
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
		log.Printf("method GetSubscription: %v", err)
		return
	}
	history, newEvents, err := eh.eventStore.GetNewEventsAndHistory(ctx, transaction.AggregateId, sub.LastSequenceID, transaction.SequenceId, commitExecutor)
	if err != nil {
		log.Printf("method GetNewEventsAndHistory: %v", err)
		return
	}
	err = eh.eventHandler.HandleEvents(ctx, history, newEvents, providerFn(transaction.AggregateId))
	if err != nil {
		log.Printf("method HandleEvents: %v", err)
		return
	}
	err = eh.eventStore.UpdateSubscription(ctx, &subscriptions.Subscription{LastSequenceID: transaction.SequenceId}, commitExecutor)
	return
}
