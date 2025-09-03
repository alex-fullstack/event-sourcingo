package consumers

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/alex-fullstack/event-sourcingo/domain/dto"
	"github.com/alex-fullstack/event-sourcingo/domain/entities"
	"github.com/alex-fullstack/event-sourcingo/domain/transactions"
	"github.com/alex-fullstack/event-sourcingo/domain/usecases/services"
	"github.com/alex-fullstack/event-sourcingo/endpoints"
	"github.com/alex-fullstack/event-sourcingo/infrastructure/postgresql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type consumer struct {
	*endpoints.Endpoint
}

func NewTransactionConsumer(ctx context.Context, ch string, conn *pgxpool.Conn, handler services.TransactionHandler, providerFn func(id uuid.UUID) entities.AggregateProvider) endpoints.EndpointStarter {
	listener := postgresql.NewListener(
		ch,
		conn,
		func(ctx context.Context, notification *pgconn.Notification) {
			tx, err := convert(notification)
			if err != nil {
				log.Printf("Transaction consumer conversation error: %v", err)
				return
			}
			err = handler.Handle(ctx, tx, providerFn)
			if err != nil {
				log.Printf("Transaction consumer handle error: %v", err)
				return
			}
		},
		func() context.Context {
			return ctx
		},
	)

	return &consumer{
		Endpoint: endpoints.NewEndpoint(
			listener.StartListen,
			listener.Shutdown,
		),
	}
}

func convert(notification *pgconn.Notification) (*transactions.Transaction, error) {
	var transactionHandle dto.TransactionHandle
	err := json.Unmarshal([]byte(notification.Payload), &transactionHandle)
	if err != nil {
		return nil, err
	}
	aggregateId, err := uuid.Parse(transactionHandle.AggregateId)
	if err != nil {
		return nil, err
	}
	transactionId, err := uuid.Parse(transactionHandle.Id)
	if err != nil {
		return nil, err
	}
	sequenceId, _ := strconv.ParseInt(transactionHandle.SequenceId, 10, 64)
	return transactions.NewTransaction(transactionId, aggregateId, sequenceId), nil
}
