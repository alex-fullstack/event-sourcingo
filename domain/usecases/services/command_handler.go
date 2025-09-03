package services

import (
	"context"

	"github.com/alex-fullstack/event-sourcingo/domain/commands"
	"github.com/alex-fullstack/event-sourcingo/domain/entities"
	"github.com/alex-fullstack/event-sourcingo/domain/events"
	"github.com/alex-fullstack/event-sourcingo/domain/usecases/repositories"
	"github.com/google/uuid"
)

type CommandHandler interface {
	Handle(
		ctx context.Context,
		cmd commands.Command,
		aggregate entities.AggregateProvider,
	) error
}

type commandHandler struct {
	store repositories.EventStore
	saver repositories.ProjectionSaver
}

func NewCommandHandler(
	store repositories.EventStore,
	saver repositories.ProjectionSaver,
) CommandHandler {
	return &commandHandler{store: store, saver: saver}
}

func (ch *commandHandler) Handle(
	ctx context.Context,
	cmd commands.Command,
	aggregate entities.AggregateProvider,
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
	if err = ch.changeAggregate(ctx, cmd, uuid.New(), aggregate, commitExecutor); err != nil {
		return err
	}
	return ch.saver.Save(ctx, aggregate.Projection())
}

func (ch *commandHandler) changeAggregate(
	ctx context.Context,
	cmd commands.Command,
	transactionID uuid.UUID,
	aggregate entities.AggregateProvider,
	commitExecutor interface{},
) error {
	history, err := ch.store.GetAggregateEvents(ctx, aggregate.ID(), commitExecutor)
	if err != nil {
		return err
	}
	if err = aggregate.Build(history); err != nil {
		return err
	}
	newEvents := make([]events.Event, len(cmd.Events))
	for i, event := range cmd.Events {
		newEvent, errNE := events.NewEvent(
			aggregate.ID(),
			transactionID,
			cmd.Type,
			aggregate.Version()+i+1,
			event.Type,
			event.Payload,
		)
		if errNE != nil {
			return errNE
		}
		newEvents[i] = newEvent
	}
	if err = aggregate.ApplyChanges(newEvents); err != nil {
		return err
	}
	return ch.store.UpdateOrCreateAggregate(ctx, transactionID, aggregate, commitExecutor)
}
