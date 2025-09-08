package services

import (
	"context"

	"github.com/alex-fullstack/event-sourcingo/domain/commands"
	"github.com/alex-fullstack/event-sourcingo/domain/entities"
	"github.com/alex-fullstack/event-sourcingo/domain/events"
	"github.com/alex-fullstack/event-sourcingo/domain/usecases/repositories"
	"github.com/google/uuid"
)

type CommandHandler[T, S, P, K any] interface {
	Handle(
		ctx context.Context,
		cmd commands.Command[T],
		aggregate entities.AggregateProvider[T, S, P, K],
	) error
}

type commandHandler[T, S, P, K, E any] struct {
	store repositories.EventStore[T, S, E]
	saver repositories.ProjectionStore[P]
}

func NewCommandHandler[T, S, P, K, E any](
	store repositories.EventStore[T, S, E],
	saver repositories.ProjectionStore[P],
) CommandHandler[T, S, P, K] {
	return &commandHandler[T, S, P, K, E]{store: store, saver: saver}
}

func (ch *commandHandler[T, S, P, K, E]) Handle(
	ctx context.Context,
	cmd commands.Command[T],
	aggregate entities.AggregateProvider[T, S, P, K],
) (err error) {
	commitExecutor, beginErr := ch.store.Begin(ctx)
	if beginErr != nil {
		return beginErr
	}
	defer func() {
		if err != nil {
			rollbackErr := ch.store.Rollback(ctx, commitExecutor)
			if rollbackErr != nil {
				err = rollbackErr
			}
		} else {
			err = ch.store.Commit(ctx, commitExecutor)
		}
	}()
	var offset int
	offset, err = ch.changeAggregate(ctx, cmd, uuid.New(), aggregate, commitExecutor)
	if err != nil {
		return err
	}
	return ch.saver.Save(ctx, aggregate.Projection(offset))
}

func (ch *commandHandler[T, S, P, K, E]) changeAggregate(
	ctx context.Context,
	cmd commands.Command[T],
	transactionID uuid.UUID,
	aggregate entities.AggregateProvider[T, S, P, K],
	commitExecutor E,
) (int, error) {
	version, payload, err := ch.store.GetLastSnapshot(ctx, aggregate.ID(), commitExecutor)
	if err != nil {
		return 0, err
	}
	var currentVersion int
	if version == 0 {
		currentVersion = -1
	} else {
		currentVersion = version
		if err = aggregate.BuildFromSnapshot(version, payload); err != nil {
			return 0, err
		}
	}

	var history []events.Event[T]
	history, err = ch.store.GetHistory(ctx, aggregate.ID(), currentVersion+1, commitExecutor)
	if err != nil {
		return 0, err
	}
	if err = aggregate.Build(history); err != nil {
		return 0, err
	}
	newEvents := make([]events.Event[T], len(cmd.Events))
	for i, event := range cmd.Events {
		newEvents[i] = events.NewEvent[T](
			aggregate.ID(),
			transactionID,
			cmd.Type,
			aggregate.Version()+i+1,
			event.Type,
			event.Payload,
		)
	}
	if err = aggregate.ApplyChanges(newEvents); err != nil {
		return 0, err
	}
	return ch.store.UpdateOrCreateAggregate(
		ctx,
		transactionID,
		aggregate,
		aggregate.Snapshot(),
		commitExecutor,
	)
}
