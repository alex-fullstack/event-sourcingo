package postgresql

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/entities"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/subscriptions"
	"strconv"
	"time"
)

type Transaction pgx.Tx

type PostgresDB struct {
	pool *pgxpool.Pool
}

func NewPostgresDB(ctx context.Context, config *pgxpool.Config) (db *PostgresDB, err error) {
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return
	}
	if err = pool.Ping(ctx); err != nil {
		return
	}
	db = &PostgresDB{pool: pool}
	return
}

func (db *PostgresDB) Acquire(ctx context.Context) (c *pgxpool.Conn, err error) {
	return db.pool.Acquire(ctx)
}

func (db *PostgresDB) Begin(ctx context.Context) (interface{}, error) {
	return db.pool.Begin(ctx)
}
func (db *PostgresDB) Commit(ctx context.Context, tx interface{}) error {
	return tx.(Transaction).Commit(ctx)
}
func (db *PostgresDB) Rollback(ctx context.Context, tx interface{}) error {
	return tx.(Transaction).Rollback(ctx)
}

func (db *PostgresDB) GetAggregateEvents(ctx context.Context, id uuid.UUID, tx interface{}) ([]events.Event, error) {
	query := `SELECT aggregate_id, transaction_id, version, command_type, event_type, payload, created_at FROM es.events WHERE aggregate_id = @id`
	args := pgx.NamedArgs{
		"id": id,
	}

	rows, err := tx.(Transaction).Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]events.Event, 0)
	for rows.Next() {
		var aggregateId, transactionId uuid.UUID
		var eventType, version, commandType int
		var payload map[string]interface{}
		var createdAt time.Time

		err = rows.Scan(&aggregateId, &transactionId, &version, &commandType, &eventType, &payload, &createdAt)
		if err != nil {
			return nil, err
		}

		result = append(result, events.Event{TransactionId: transactionId, AggregateId: aggregateId, CommandType: commandType, Type: eventType, Version: version, Payload: payload, CreatedAt: &createdAt})
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (db *PostgresDB) GetNewEventsAndHistory(ctx context.Context, id uuid.UUID, firstSequenceId, lastSequenceId int64, tx interface{}) ([]events.Event, []events.Event, error) {
	query := `SELECT sequence_id::text, e.aggregate_id, transaction_id, version, command_type, event_type, payload, created_at FROM es.transactions AS t
	JOIN es.events AS e
	ON e.transaction_id=t.id
	WHERE sequence_id <= @lastSequenceId::xid8 AND t.aggregate_id=@aggregateId
	ORDER BY sequence_id`
	args := pgx.NamedArgs{
		"lastSequenceId": lastSequenceId,
		"aggregateId":    id,
	}
	var history, newEvents []events.Event
	rows, err := tx.(Transaction).Query(ctx, query, args)
	if err != nil {
		return history, newEvents, err
	}
	defer rows.Close()

	for rows.Next() {
		var sequenceId string
		var aggregateId, transactionId uuid.UUID
		var eventType, version, commandType int
		var payload map[string]interface{}
		var createdAt time.Time

		err = rows.Scan(&sequenceId, &aggregateId, &transactionId, &version, &commandType, &eventType, &payload, &createdAt)
		if err != nil {
			return []events.Event{}, []events.Event{}, err
		}
		seqId, err := strconv.ParseInt(sequenceId, 10, 64)
		if err != nil {
			return []events.Event{}, []events.Event{}, err
		}
		event := events.Event{TransactionId: transactionId, AggregateId: aggregateId, CommandType: commandType, Type: eventType, Version: version, Payload: payload, CreatedAt: &createdAt}
		if seqId <= firstSequenceId {
			history = append(history, event)
		} else {
			newEvents = append(newEvents, event)
		}
	}
	if err = rows.Err(); err != nil {
		return []events.Event{}, []events.Event{}, err
	}
	return history, newEvents, nil
}

func (db *PostgresDB) UpdateOrCreateAggregate(ctx context.Context, transactionId uuid.UUID, reader entities.AggregateReader, tx interface{}) (err error) {
	currentVersion, nextVersion := reader.BaseVersion(), reader.Version()
	if currentVersion == 0 {
		err = db.createVersion(ctx, reader.ID(), nextVersion, tx.(Transaction))
	} else {
		err = db.updateVersion(ctx, reader.ID(), currentVersion, nextVersion, tx.(Transaction))
	}
	if err != nil {
		return
	}
	err = db.insertEvents(ctx, reader.Changes(), tx.(Transaction))
	if err != nil {
		return
	}
	err = db.insertTransaction(ctx, transactionId, reader.ID(), tx.(Transaction))
	return
}

func (db *PostgresDB) Close() {
	db.pool.Close()
}

func (db *PostgresDB) GetSubscription(ctx context.Context, tx interface{}) (*subscriptions.Subscription, error) {
	query := `SELECT last_sequence_id::text FROM es.subscription WHERE id = 1 FOR UPDATE SKIP LOCKED`

	var lastSequenceId string
	err := tx.(Transaction).QueryRow(ctx, query).Scan(&lastSequenceId)
	if err != nil {
		return nil, err
	}
	sequenceId, err := strconv.ParseInt(lastSequenceId, 10, 64)
	if err != nil {
		return nil, err
	}
	return &subscriptions.Subscription{LastSequenceID: sequenceId}, nil
}
func (db *PostgresDB) UpdateSubscription(ctx context.Context, sub *subscriptions.Subscription, tx interface{}) error {
	query := `UPDATE es.subscription SET last_sequence_id = @lastSequenceId::xid8  WHERE id = 1`
	args := pgx.NamedArgs{
		"lastSequenceId": sub.LastSequenceID,
	}
	_, err := tx.(Transaction).Exec(ctx, query, args)
	return err
}

func (db *PostgresDB) createVersion(ctx context.Context, id uuid.UUID, version int, tx Transaction) error {
	query := `INSERT INTO es.aggregates (id, version) VALUES (@id, @version)`
	args := pgx.NamedArgs{
		"id":      id,
		"version": version,
	}
	_, err := tx.Exec(ctx, query, args)
	return err
}
func (db *PostgresDB) updateVersion(ctx context.Context, id uuid.UUID, currentVersion, nextVersion int, tx Transaction) error {
	query := `UPDATE es.aggregates SET version = @nextVersion WHERE id = @id AND version = @currentVersion`
	args := pgx.NamedArgs{
		"id":             id,
		"currentVersion": currentVersion,
		"nextVersion":    nextVersion,
	}
	_, err := tx.Exec(ctx, query, args)
	return err
}

func (db *PostgresDB) insertEvents(ctx context.Context, events []events.Event, tx Transaction) (err error) {
	query := `INSERT INTO es.events (aggregate_id, transaction_id, version, command_type, event_type, payload) VALUES (@aggregateId, @transactionId, @version, @commandType, @eventType, @payload)`

	batch := &pgx.Batch{}
	for _, event := range events {
		args := pgx.NamedArgs{
			"aggregateId":   event.AggregateId,
			"version":       event.Version,
			"transactionId": event.TransactionId,
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

	for _ = range events {
		_, err = results.Exec()
		if err != nil {
			return
		}
	}
	return
}

func (db *PostgresDB) insertTransaction(ctx context.Context, id, aggregateId uuid.UUID, tx Transaction) (err error) {
	query := `INSERT INTO es.transactions (id, aggregate_id) VALUES (@id, @aggregateId)`
	args := pgx.NamedArgs{
		"id":          id,
		"aggregateId": aggregateId,
	}
	_, err = tx.Exec(ctx, query, args)
	return
}
