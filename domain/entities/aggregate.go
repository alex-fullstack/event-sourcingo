package entities

import (
	"errors"

	"github.com/alex-fullstack/event-sourcingo/domain/events"
	"github.com/google/uuid"
)

type AggregateReader[T any] interface {
	ID() uuid.UUID
	Cap() int
	Version() int
	BaseVersion() int
	Changes() []events.Event[T]
}

type AggregateWriter[T, S any] interface {
	Build(events []events.Event[T]) error
	BuildFromSnapshot(version int, payload S) error
	ApplyChanges(events []events.Event[T]) error
	ApplyChange(e events.Event[T]) error
}

type AggregateProvider[T, S, P, K any] interface {
	AggregateReader[T]
	AggregateWriter[T, S]
	Snapshot() S
	Projection() P
	IntegrationEvent(evType int) events.IntegrationEvent[K]
}

type Aggregate[T, S any] struct {
	id            uuid.UUID
	cap           int
	version       int
	baseVersion   int
	changes       []events.Event[T]
	apply         func(events.Event[T]) error
	applySnapshot func(payload S) error
}

func NewAggregate[T, S any](
	id uuid.UUID,
	capacity int,
	apply func(events.Event[T]) error,
	applySnapshot func(payload S) error,
) *Aggregate[T, S] {
	return &Aggregate[T, S]{
		id:            id,
		cap:           capacity,
		apply:         apply,
		applySnapshot: applySnapshot,
		changes:       make([]events.Event[T], 0),
	}
}

func (a *Aggregate[T, S]) Changes() []events.Event[T] {
	return a.changes
}

func (a *Aggregate[T, S]) ID() uuid.UUID {
	return a.id
}

func (a *Aggregate[T, S]) Cap() int {
	return a.cap
}

func (a *Aggregate[T, S]) Version() int {
	return a.version
}

func (a *Aggregate[T, S]) BaseVersion() int {
	return a.baseVersion
}

func (a *Aggregate[T, S]) ApplyChanges(events []events.Event[T]) error {
	for _, event := range events {
		err := a.ApplyChange(event)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Aggregate[T, S]) ApplyChange(e events.Event[T]) error {
	if e.Version != a.version+1 {
		return errors.New("cannot add an event with an invalid version")
	}
	err := a.apply(e)
	if err != nil {
		return err
	}
	a.changes = append(a.changes, e)
	a.version = e.Version
	return nil
}

func (a *Aggregate[T, S]) Build(events []events.Event[T]) error {
	for _, event := range events {
		if event.Version <= a.version {
			return errors.New("cannot load an event with an invalid version")
		}
		err := a.apply(event)
		if err != nil {
			return err
		}
		a.version = event.Version
		a.baseVersion = event.Version
	}
	return nil
}

func (a *Aggregate[T, S]) BuildFromSnapshot(version int, payload S) error {
	if version <= a.version {
		return errors.New("cannot load an event with an invalid version")
	}
	err := a.applySnapshot(payload)
	if err != nil {
		return err
	}
	a.version = version
	a.baseVersion = version
	return nil
}
