package consumers

import (
	"context"
	"encoding/json"
	"log/slog"
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

func NewTransactionConsumer(
	ctx context.Context,
	ch string,
	conn *pgxpool.Conn,
	handler services.TransactionHandler,
	providerFn func(id uuid.UUID) entities.AggregateProvider,
) endpoints.EndpointStarter {
	logger := slog.Default()
	listener := postgresql.NewListener(
		ch,
		conn,
		func(ctx context.Context, notification *pgconn.Notification) {
			tx, err := convert(notification)
			if err != nil {
				slog.Error(err.Error())
				return
			}
			err = handler.Handle(ctx, tx, providerFn)
			if err != nil {
				slog.Error(err.Error())
				return
			}
		},
		func() context.Context {
			return ctx
		},
		logger,
	)

	return &consumer{
		Endpoint: endpoints.NewEndpoint(
			listener.StartListen,
			listener.Shutdown,
			logger,
		),
	}
}

func convert(notification *pgconn.Notification) (*transactions.Transaction, error) {
	var transactionHandle dto.TransactionHandle
	err := json.Unmarshal([]byte(notification.Payload), &transactionHandle)
	if err != nil {
		return nil, err
	}
	aggregateID, err := uuid.Parse(transactionHandle.AggregateID)
	if err != nil {
		return nil, err
	}
	transactionID, err := uuid.Parse(transactionHandle.ID)
	if err != nil {
		return nil, err
	}
	sequenceID, _ := strconv.ParseInt(transactionHandle.SequenceID, 10, 64)
	return transactions.NewTransaction(transactionID, aggregateID, sequenceID), nil
}
