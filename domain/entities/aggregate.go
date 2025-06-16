package entities

import (
	"errors"

	"github.com/google/uuid"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
)

type AggregateReader interface {
	ID() uuid.UUID
	Version() int
	BaseVersion() int
	Changes() []events.Event
	Projection() interface{}
	IntegrationEvent(evType int) (events.IntegrationEvent, error)
}

type AggregateWriter interface {
	Build(events []events.Event) error
	ApplyChanges(events []events.Event) error
}

type AggregateProvider interface {
	AggregateReader
	AggregateWriter
}

type Aggregate struct {
	id          uuid.UUID
	version     int
	baseVersion int
	changes     []events.Event
	apply       func(events.Event) error
}

func NewAggregate(id uuid.UUID, apply func(events.Event) error) *Aggregate {
	return &Aggregate{id: id, apply: apply, changes: make([]events.Event, 0)}
}

func (a *Aggregate) Changes() []events.Event {
	return a.changes
}

func (a *Aggregate) ID() uuid.UUID {
	return a.id
}

func (a *Aggregate) Version() int {
	return a.version
}

func (a *Aggregate) BaseVersion() int {
	return a.baseVersion
}

func (a *Aggregate) ApplyChanges(events []events.Event) error {
	for _, event := range events {
		err := a.applyChange(event)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Aggregate) applyChange(e events.Event) error {
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

func (a *Aggregate) Build(events []events.Event) error {
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
