package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	AggregateId   uuid.UUID
	TransactionId uuid.UUID
	CommandType   int
	Version       int
	Payload       map[string]interface{}
	Type          int
	CreatedAt     *time.Time
}

type IntegrationEvent struct {
	Id      string
	Type    int
	Payload map[string]interface{}
}

func NewEvent(aggregateId, transactionId uuid.UUID, cmdType int, version, evType int, rawPayload interface{}) (Event, error) {
	payload, err := payloadFromRaw(rawPayload)
	if err != nil {
		return Event{}, err
	}
	return Event{TransactionId: transactionId, CommandType: cmdType, AggregateId: aggregateId, Version: version, Type: evType, Payload: payload}, nil
}

func NewIntegrationEvent(id uuid.UUID, evType int, rawPayload interface{}) (IntegrationEvent, error) {
	payload, err := payloadFromRaw(rawPayload)
	if err != nil {
		return IntegrationEvent{}, err
	}

	return IntegrationEvent{Id: id.String(), Type: evType, Payload: payload}, nil
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
