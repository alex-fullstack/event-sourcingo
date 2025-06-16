package services

import (
	"context"
	"github.com/google/uuid"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/commands"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/entities"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/usecases/repositories"
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

func NewCommandHandler(store repositories.EventStore, saver repositories.ProjectionSaver) CommandHandler {
	return &commandHandler{store: store, saver: saver}
}

func (ch *commandHandler) Handle(
	ctx context.Context,
	cmd commands.Command,
	aggregate entities.AggregateProvider,
) (err error) {
	commitExecutor, beginErr := ch.store.Begin(ctx)
	if beginErr != nil {
		err = beginErr
		return
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
		return
	}
	err = ch.saver.Save(ctx, aggregate.Projection())
	return
}

func (ch *commandHandler) changeAggregate(ctx context.Context, cmd commands.Command, transactionId uuid.UUID, aggregate entities.AggregateProvider, commitExecutor interface{}) error {
	history, err := ch.store.GetAggregateEvents(ctx, aggregate.ID(), commitExecutor)
	if err != nil {
		return err
	}
	if err = aggregate.Build(history); err != nil {
		return err
	}
	newEvents := make([]events.Event, len(cmd.Events))
	for i, event := range cmd.Events {
		newEvent, err := events.NewEvent(aggregate.ID(), transactionId, cmd.Type, aggregate.Version()+i+1, event.Type, event.Payload)
		if err != nil {
			return err
		}
		newEvents[i] = newEvent
	}
	if err = aggregate.ApplyChanges(newEvents); err != nil {
		return err
	}
	return ch.store.UpdateOrCreateAggregate(ctx, transactionId, aggregate, commitExecutor)
}
