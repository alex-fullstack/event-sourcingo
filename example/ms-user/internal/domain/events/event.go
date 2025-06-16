package events

import (
	"time"
	"user/internal/domain/dto"

	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
)

type UserEventType int

const (
	CredentialsCreated UserEventType = iota + 1
	UserSigned
	UserAuthenticated
	ConfirmationCreated
	ConfirmationUpdated
	EmailConfirmed
	UserConfirmationRequested
	UserConfirmationRejected
	UserConfirmationCompleted
)

func (uet UserEventType) String() string {
	return [...]string{
		"credentials_created",
		"user_signed",
		"user_authenticated",
		"confirmation_created",
		"confirmation_updated",
		"email_verified",
		"user_confirmation_requested",
		"user_confirmation_rejected",
		"user_confirmation_completed",
	}[uet-1]
}

func (uet UserEventType) Index() int {
	return int(uet)
}

type UserHistoryRecord struct {
	ActivityType string    `json:"activity_type"`
	Timestamp    time.Time `json:"timestamp"`
	Device       string    `json:"device"`
}

type IntegrationEventPayload struct {
	Email          string              `json:"email"`
	EmailVerified  bool                `json:"email_verified"`
	ActivationCode *string             `json:"activation_code"`
	History        []UserHistoryRecord `json:"history"`
}

func NewUserHistoryRecord(record dto.ActivityRecord) UserHistoryRecord {
	return UserHistoryRecord{
		ActivityType: record.Type,
		Timestamp:    record.Timestamp,
		Device:       record.Device,
	}
}

func NewIntegrationEventPayload(projection dto.UserProjection) IntegrationEventPayload {
	payload := IntegrationEventPayload{}
	var history = make([]UserHistoryRecord, len(projection.History))
	for i, historyRecord := range projection.History {
		history[i] = NewUserHistoryRecord(historyRecord)
	}
	payload.History = history
	if projection.Credentials != nil {
		payload.Email = projection.Credentials.Email
		payload.EmailVerified = projection.Credentials.Confirmed
	}
	if projection.Confirmation != nil {
		payload.ActivationCode = &projection.Confirmation.Code
	}
	return payload
}

func NewIntegrationEvent(evType int, projection dto.UserProjection) (events.IntegrationEvent, error) {
	return events.NewIntegrationEvent(projection.ID, evType, NewIntegrationEventPayload(projection))
}
