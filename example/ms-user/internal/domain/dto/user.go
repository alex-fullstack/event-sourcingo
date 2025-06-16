package dto

import (
	"github.com/google/uuid"
)

type UserSign struct {
	Credentials CredentialsCreate
	Activity    ActivityInput
}

type UserLogin struct {
	Credentials AuthInput
	Activity    ActivityInput
}

type UserRequestConfirmation struct {
	UserID   uuid.UUID
	Activity ActivityInput
}

type UserProjection struct {
	ID           uuid.UUID               `bson:"id"`
	Credentials  *CredentialsProjection  `bson:"credentials"`
	Confirmation *ConfirmationProjection `bson:"confirmation"`
	History      []ActivityRecord        `bson:"history"`
}

func NewUserSign(credentials CredentialsCreate, activity ActivityInput) UserSign {
	return UserSign{
		Credentials: credentials,
		Activity:    activity,
	}
}

func NewUserLogin(credentials AuthInput, activity ActivityInput) UserLogin {
	return UserLogin{
		Credentials: credentials,
		Activity:    activity,
	}
}

func NewUserProjection(
	id uuid.UUID,
	credentials *CredentialsProjection,
	confirmation *ConfirmationProjection,
	history []ActivityRecord,
) UserProjection {
	return UserProjection{
		ID:           id,
		Credentials:  credentials,
		Confirmation: confirmation,
		History:      history,
	}
}
