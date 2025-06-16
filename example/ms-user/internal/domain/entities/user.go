package entities

import (
	"errors"
	"fmt"
	"user/internal/domain/dto"
	"user/internal/domain/events"

	"github.com/google/uuid"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/entities"
	coreEvents "gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/events"
)

type User struct {
	*entities.Aggregate
	credentials  *Credentials
	confirmation *Confirmation
	activities   []*Activity
}

func NewUser(id uuid.UUID) *User {
	res := &User{
		activities: make([]*Activity, 0),
	}
	res.Aggregate = entities.NewAggregate(id, func(event coreEvents.Event) error {
		payload := event.Payload
		var err error
		var activity *Activity
		switch events.UserEventType(event.Type) {
		case events.UserSigned:
			activity, err = NewActivity(Registration, payload)
		case events.CredentialsCreated:
			res.credentials, err = NewCredentials(payload)
		case events.UserAuthenticated:
			activity, err = NewActivity(Authentication, payload)
		case events.ConfirmationCreated, events.ConfirmationUpdated:
			res.confirmation, err = NewConfirmation(payload)
		case events.EmailConfirmed:
			res.credentials.Confirmed = true
			res.confirmation = nil
		case events.UserConfirmationCompleted, events.UserConfirmationRequested, events.UserConfirmationRejected:
			activity, err = NewActivity(Verification, payload)
		default:
			err = errors.New("unhandled default case")
		}
		if err != nil {
			return err
		}
		if activity != nil {
			res.activities = append(res.activities, activity)
		}
		return nil
	})

	return res
}

func (u *User) Projection() interface{} {
	history := make([]dto.ActivityRecord, len(u.activities))
	for i, activity := range u.activities {
		history[i] = dto.NewActivityRecord(activity.Timestamp, activity.Device, fmt.Sprint(activity.Type))
	}
	var credentials *dto.CredentialsProjection
	var confirmation *dto.ConfirmationProjection
	if u.credentials != nil {
		credentials = dto.NewCredentialsProjection(u.credentials.Email, u.credentials.PasswordHash, u.credentials.Confirmed)
	}
	if u.confirmation != nil {
		confirmation = dto.NewConfirmationProjection(u.confirmation.Code, u.confirmation.Expired)
	}
	return dto.UserProjection{
		ID:           u.ID(),
		Credentials:  credentials,
		Confirmation: confirmation,
		History:      history,
	}
}

func (u *User) IntegrationEvent(evType int) (coreEvents.IntegrationEvent, error) {
	projection, ok := u.Projection().(dto.UserProjection)
	if !ok {
		return coreEvents.IntegrationEvent{}, errors.New("wrong data type, need dto.UserProjection")
	}
	return events.NewIntegrationEvent(evType, projection)
}
