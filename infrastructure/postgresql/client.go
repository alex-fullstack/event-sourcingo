package postgresql

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/alex-fullstack/event-sourcingo/domain/entities"
	"github.com/alex-fullstack/event-sourcingo/domain/events"
	"github.com/alex-fullstack/event-sourcingo/domain/subscriptions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	cfg                        *pgxpool.Config
	snapshotEventMultiplicator int
}

type Transaction pgx.Tx

type PostgresDB[T, S any] struct {
	pool                       *pgxpool.Pool
	snapshotEventMultiplicator int
}

func NewConfig(cfg *pgxpool.Config, snapshotEventMultiplicator int) Config {
	return Config{cfg: cfg, snapshotEventMultiplicator: snapshotEventMultiplicator}
}

func NewPostgresDB[T, S any](
	ctx context.Context,
	config Config,
) (*PostgresDB[T, S], error) {
	pool, err := pgxpool.NewWithConfig(ctx, config.cfg)
	if err != nil {
		return nil, err
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &PostgresDB[T, S]{
		pool:                       pool,
		snapshotEventMultiplicator: config.snapshotEventMultiplicator,
	}, nil
}

func (db *PostgresDB[T, S]) Acquire(
	ctx context.Context,
) (*pgxpool.Conn, error) {
	return db.pool.Acquire(ctx)
}

func (db *PostgresDB[T, S]) Begin(
	ctx context.Context,
) (Transaction, error) {
	return db.pool.Begin(ctx)
}

func (db *PostgresDB[T, S]) Commit(
	ctx context.Context,
	tx Transaction,
) error {
	return tx.Commit(ctx)
}

func (db *PostgresDB[T, S]) Rollback(
	ctx context.Context,
	tx Transaction,
) error {
	return tx.Rollback(ctx)
}

func (db *PostgresDB[T, S]) GetLastSnapshot(
	ctx context.Context,
	id uuid.UUID,
	tx Transaction,
) (int, S, error) {
	query := `SELECT version, payload FROM es.snapshots WHERE aggregate_id = @id ORDER BY version DESC LIMIT 1`
	args := pgx.NamedArgs{
		"id": id,
	}
	var payload S
	var version int
	rows, err := tx.Query(ctx, query, args)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return version, payload, nil
		}
		return version, payload, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&version,
			&payload,
		)
		if err != nil {
			return version, payload, err
		}
	}
	if err = rows.Err(); err != nil {
		return version, payload, err
	}

	return version, payload, nil
}

func (db *PostgresDB[T, S]) GetHistory(
	ctx context.Context,
	id uuid.UUID,
	fromVersion int,
	tx Transaction,
) ([]events.Event[T], error) {
	query := `SELECT aggregate_id, transaction_id, version, command_type, event_type, payload, created_at FROM es.events WHERE aggregate_id = @id AND version >= @fromVersion` //nolint:lll
	args := pgx.NamedArgs{
		"id":          id,
		"fromVersion": fromVersion,
	}
	rows, err := tx.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]events.Event[T], 0)
	for rows.Next() {
		var aggregateID, transactionID uuid.UUID
		var eventType, version, commandType int
		var payload T
		var createdAt time.Time

		err = rows.Scan(
			&aggregateID,
			&transactionID,
			&version,
			&commandType,
			&eventType,
			&payload,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(
			result,
			events.Event[T]{
				TransactionID: transactionID,
				AggregateID:   aggregateID,
				CommandType:   commandType,
				Type:          eventType,
				Version:       version,
				Payload:       payload,
				CreatedAt:     &createdAt,
			},
		)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (db *PostgresDB[T, S]) GetNewEventsAndHistory(
	ctx context.Context,
	id uuid.UUID,
	firstSequenceID, lastSequenceID int64,
	tx Transaction,
) ([]events.Event[T], []events.Event[T], error) {
	query := `SELECT sequence_id::text, e.aggregate_id, transaction_id, version, command_type, event_type, payload, created_at FROM es.transactions AS t JOIN es.events AS e ON e.transaction_id=t.id WHERE sequence_id <= @lastSequenceId::xid8 AND t.aggregate_id=@aggregateId ORDER BY sequence_id` //nolint:lll
	args := pgx.NamedArgs{
		"lastSequenceId": lastSequenceID,
		"aggregateId":    id,
	}
	var history, newEvents []events.Event[T]
	rows, err := tx.Query(ctx, query, args)
	if err != nil {
		return history, newEvents, err
	}
	defer rows.Close()

	for rows.Next() {
		var sequenceID string
		var aggregateID, transactionID uuid.UUID
		var eventType, version, commandType int
		var payload T
		var createdAt time.Time

		err = rows.Scan(
			&sequenceID,
			&aggregateID,
			&transactionID,
			&version,
			&commandType,
			&eventType,
			&payload,
			&createdAt,
		)
		if err != nil {
			return []events.Event[T]{}, []events.Event[T]{}, err
		}
		seqID, errParse := strconv.ParseInt(sequenceID, 10, 64)
		if errParse != nil {
			return []events.Event[T]{}, []events.Event[T]{}, errParse
		}
		event := events.Event[T]{
			TransactionID: transactionID,
			AggregateID:   aggregateID,
			CommandType:   commandType,
			Type:          eventType,
			Version:       version,
			Payload:       payload,
			CreatedAt:     &createdAt,
		}
		if seqID <= firstSequenceID {
			history = append(history, event)
		} else {
			newEvents = append(newEvents, event)
		}
	}
	if err = rows.Err(); err != nil {
		return []events.Event[T]{}, []events.Event[T]{}, err
	}
	return history, newEvents, nil
}

func (db *PostgresDB[T, S]) UpdateOrCreateAggregate(
	ctx context.Context,
	transactionID uuid.UUID,
	reader entities.AggregateReader[T],
	snapshot S,
	tx Transaction,
) (int, error) {
	currentVersion, nextVersion := reader.BaseVersion(), reader.Version()
	var err error
	if currentVersion == 0 {
		err = db.createVersion(ctx, reader.ID(), nextVersion, tx)
	} else {
		err = db.updateVersion(ctx, reader.ID(), currentVersion, nextVersion, tx)
	}
	if err != nil {
		return 0, err
	}
	if nextVersion/db.snapshotEventMultiplicator > currentVersion/db.snapshotEventMultiplicator {
		err = db.insertSnapshot(ctx, reader.ID(), nextVersion, snapshot, tx)
	}
	if err != nil {
		return 0, err
	}
	err = db.insertEvents(ctx, reader.Changes(), tx)
	if err != nil {
		return 0, err
	}
	return nextVersion / db.snapshotEventMultiplicator, db.insertTransaction(
		ctx,
		transactionID,
		reader.ID(),
		tx,
	)
}

func (db *PostgresDB[T, S]) Close() {
	db.pool.Close()
}

func (db *PostgresDB[T, S]) GetSubscription(
	ctx context.Context,
	tx Transaction,
) (*subscriptions.Subscription, error) {
	query := `SELECT last_sequence_id::text FROM es.subscription WHERE id = 1 FOR UPDATE SKIP LOCKED`

	var lastSequenceID string
	err := tx.QueryRow(ctx, query).Scan(&lastSequenceID)
	if err != nil {
		return nil, err
	}
	sequenceID, err := strconv.ParseInt(lastSequenceID, 10, 64)
	if err != nil {
		return nil, err
	}
	return &subscriptions.Subscription{LastSequenceID: sequenceID}, nil
}

func (db *PostgresDB[T, S]) UpdateSubscription(
	ctx context.Context,
	sub *subscriptions.Subscription,
	tx Transaction,
) error {
	query := `UPDATE es.subscription SET last_sequence_id = @lastSequenceId::xid8  WHERE id = 1`
	args := pgx.NamedArgs{
		"lastSequenceId": sub.LastSequenceID,
	}
	_, err := tx.Exec(ctx, query, args)
	return err
}

func (db *PostgresDB[T, S]) createVersion(
	ctx context.Context,
	id uuid.UUID,
	version int,
	tx Transaction,
) error {
	query := `INSERT INTO es.aggregates (id, version) VALUES (@id, @version)`
	args := pgx.NamedArgs{
		"id":      id,
		"version": version,
	}
	_, err := tx.Exec(ctx, query, args)
	return err
}

func (db *PostgresDB[T, S]) updateVersion(
	ctx context.Context,
	id uuid.UUID,
	currentVersion, nextVersion int,
	tx Transaction,
) error {
	query := `UPDATE es.aggregates SET version = @nextVersion WHERE id = @id AND version = @currentVersion`
	args := pgx.NamedArgs{
		"id":             id,
		"currentVersion": currentVersion,
		"nextVersion":    nextVersion,
	}
	_, err := tx.Exec(ctx, query, args)
	return err
}

func (db *PostgresDB[T, S]) insertEvents(
	ctx context.Context,
	events []events.Event[T],
	tx Transaction,
) (err error) {
	query := `INSERT INTO es.events (aggregate_id, transaction_id, version, command_type, event_type, payload) VALUES (@aggregateId, @transactionId, @version, @commandType, @eventType, @payload)` //nolint:lll

	batch := &pgx.Batch{}
	for _, event := range events {
		args := pgx.NamedArgs{
			"aggregateId":   event.AggregateID,
			"version":       event.Version,
			"transactionId": event.TransactionID,
			"eventType":     event.Type,
			"commandType":   event.CommandType,
			"payload":       event.Payload,
		}
		batch.Queue(query, args)
	}

	results := tx.SendBatch(ctx, batch)
	defer func() {
		closeErr := results.Close()
		if err == nil {
			err = closeErr
		}
	}()

	for range events {
		_, err = results.Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *PostgresDB[T, S]) insertTransaction(
	ctx context.Context,
	id, aggregateID uuid.UUID,
	tx Transaction,
) error {
	query := `INSERT INTO es.transactions (id, aggregate_id) VALUES (@id, @aggregateId)`
	args := pgx.NamedArgs{
		"id":          id,
		"aggregateId": aggregateID,
	}
	_, err := tx.Exec(ctx, query, args)
	return err
}

func (db *PostgresDB[T, S]) insertSnapshot(
	ctx context.Context,
	aggregateID uuid.UUID,
	version int,
	payload S,
	tx Transaction,
) error {
	query := `INSERT INTO es.snapshots (aggregate_id, version, payload) VALUES (@aggregateId, @version, @payload)`
	args := pgx.NamedArgs{
		"aggregateId": aggregateID,
		"version":     version,
		"payload":     payload,
	}
	_, err := tx.Exec(ctx, query, args)
	return err
}
