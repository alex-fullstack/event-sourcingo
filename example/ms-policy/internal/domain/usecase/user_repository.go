package usecase

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error)
}
