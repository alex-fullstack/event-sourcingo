package dto

import (
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
)

type UserHistoryRecord struct {
	ActivityType string    `json:"activity_type"`
	Timestamp    time.Time `json:"timestamp"`
	Device       string    `json:"device"`
}

func NewUserHistoryRecord(payload map[string]interface{}) (UserHistoryRecord, error) {
	uhr := &UserHistoryRecord{}
	return uhr.parse(payload)
}

func (uhr *UserHistoryRecord) parse(payload map[string]interface{}) (UserHistoryRecord, error) {
	activityType, ok := payload["activity_type"].(string)
	if !ok {
		return UserHistoryRecord{}, errors.New("activity_type could not be parsed")
	}
	uhr.ActivityType = activityType
	timestamp, ok := payload["timestamp"].(string)
	if !ok {
		return UserHistoryRecord{}, errors.New("timestamp could not be parsed")
	}
	var err error
	uhr.Timestamp, err = time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return UserHistoryRecord{}, err
	}
	device, ok := payload["device"].(string)
	if !ok {
		return UserHistoryRecord{}, errors.New("device could not be parsed")
	}
	uhr.Device = device
	return *uhr, nil
}

type UserUpsert struct {
	Email          string              `json:"email"`
	ActivationCode *string             `json:"activation_code"`
	EmailVerified  bool                `json:"email_verified"`
	History        []UserHistoryRecord `json:"history"`
}

func NewUserUpsert(payload map[string]interface{}) (UserUpsert, error) {
	uu := &UserUpsert{History: make([]UserHistoryRecord, 0)}
	return uu.parse(payload)
}

func (uu *UserUpsert) parse(payload map[string]interface{}) (UserUpsert, error) {
	email, ok := payload["email"].(string)
	if !ok {
		return UserUpsert{}, errors.New("email could not be parsed")
	}
	uu.Email = email
	rawActivationCode := payload["activation_code"]
	if rawActivationCode != nil {
		activationCode, codeOk := rawActivationCode.(string)
		if !codeOk {
			return UserUpsert{}, errors.New("activation_code could not be parsed")
		}
		uu.ActivationCode = &activationCode
	}
	emailVerified, ok := payload["email_verified"].(bool)
	if !ok {
		return UserUpsert{}, errors.New("email_verified could not be parsed")
	}
	uu.EmailVerified = emailVerified
	history, ok := payload["history"].([]interface{})
	if !ok {
		return UserUpsert{}, errors.New("history could not be parsed")
	}
	if len(history) > 0 {
		uu.History = slices.Collect(func(yield func(UserHistoryRecord) bool) {
			for _, rawItem := range history {
				item, _ := rawItem.(map[string]interface{})
				rec, _ := NewUserHistoryRecord(item)
				if !yield(rec) {
					return
				}
			}
		})
	}

	return *uu, nil
}

type UserPolicy struct {
	ID          uuid.UUID    `json:"id"`
	Role        *Role        `json:"role"`
	Permissions []Permission `json:"permissions"`
}

type UserDocument struct {
	ID       uuid.UUID    `json:"id"`
	Info     UserUpsert   `json:"info"`
	Policies []UserPolicy `json:"policies"`
}
