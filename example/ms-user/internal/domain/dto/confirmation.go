package dto

import (
	"time"

	"github.com/google/uuid"
)

type ConfirmationProjection struct {
	Code    string
	Expired time.Time
}

type ConfirmationInput struct {
	Code    string    `json:"code"`
	Expired time.Time `json:"expired"`
}

type ConfirmInput struct {
	Code     string    `json:"code"`
	UserID   uuid.UUID `json:"user_id"`
	Activity ActivityInput
}

func NewConfirmationProjection(code string, expired time.Time) *ConfirmationProjection {
	return &ConfirmationProjection{Code: code, Expired: expired}
}

func NewConfirmationInput(code string, expired time.Time) ConfirmationInput {
	return ConfirmationInput{Code: code, Expired: expired}
}
