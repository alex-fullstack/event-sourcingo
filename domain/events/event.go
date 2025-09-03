package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	AggregateID   uuid.UUID
	TransactionID uuid.UUID
	CommandType   int
	Version       int
	Payload       map[string]interface{}
	Type          int
	CreatedAt     *time.Time
}

type IntegrationEvent struct {
	ID      string                 `json:"id,omitempty"`
	Type    int                    `json:"type,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

func NewEvent(
	aggregateID, transactionID uuid.UUID,
	cmdType int,
	version, evType int,
	rawPayload interface{},
) (Event, error) {
	payload, err := payloadFromRaw(rawPayload)
	if err != nil {
		return Event{}, err
	}
	return Event{
		TransactionID: transactionID,
		CommandType:   cmdType,
		AggregateID:   aggregateID,
		Version:       version,
		Type:          evType,
		Payload:       payload,
	}, nil
}

func NewIntegrationEvent(
	id uuid.UUID,
	evType int,
	rawPayload interface{},
) (IntegrationEvent, error) {
	payload, err := payloadFromRaw(rawPayload)
	if err != nil {
		return IntegrationEvent{}, err
	}

	return IntegrationEvent{ID: id.String(), Type: evType, Payload: payload}, nil
}

func payloadFromRaw(rawPayload interface{}) (map[string]interface{}, error) {
	var jsonData []byte
	jsonData, err := json.Marshal(rawPayload)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	err = json.Unmarshal(jsonData, &payload)
	return payload, err
}
