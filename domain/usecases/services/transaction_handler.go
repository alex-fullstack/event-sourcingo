package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/entities"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/subscriptions"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/transactions"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/usecases/repositories"
	"log"
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
		log.Println(fmt.Sprintf("method GetSubscription: %v", err))
		return
	}
	history, newEvents, err := eh.eventStore.GetNewEventsAndHistory(ctx, transaction.AggregateId, sub.LastSequenceID, transaction.SequenceId, commitExecutor)
	if err != nil {
		log.Println(fmt.Sprintf("method GetNewEventsAndHistory: %v", err))
		return
	}
	err = eh.eventHandler.HandleEvents(ctx, history, newEvents, providerFn(transaction.AggregateId))
	if err != nil {
		log.Println(fmt.Sprintf("method HandleEvents: %v", err))
		return
	}
	err = eh.eventStore.UpdateSubscription(ctx, &subscriptions.Subscription{LastSequenceID: transaction.SequenceId}, commitExecutor)
	return
}
