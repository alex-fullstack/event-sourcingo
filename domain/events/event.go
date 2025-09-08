package events

import (
	"time"

	"github.com/google/uuid"
)

type Event[T any] struct {
	AggregateID   uuid.UUID
	TransactionID uuid.UUID
	CommandType   int
	Version       int
	Payload       T
	Type          int
	CreatedAt     *time.Time
}

type IntegrationEvent[T any] struct {
	ID      string `json:"id"`
	Type    int    `json:"type"`
	Payload T      `json:"payload"`
}

func NewEvent[T any](
	aggregateID, transactionID uuid.UUID,
	cmdType int,
	version, evType int,
	payload T,
) Event[T] {
	return Event[T]{
		TransactionID: transactionID,
		CommandType:   cmdType,
		AggregateID:   aggregateID,
		Version:       version,
		Type:          evType,
		Payload:       payload,
	}
}

func NewIntegrationEvent[T any](
	id uuid.UUID,
	evType int,
	payload T,
) IntegrationEvent[T] {
	return IntegrationEvent[T]{
		ID:      id.String(),
		Type:    evType,
		Payload: payload,
	}
}
