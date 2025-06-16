package usecase

import (
	"context"

	"github.com/google/uuid"
)

type PolicyRepository interface {
	CheckRole(ctx context.Context, userID uuid.UUID, code string) bool
}
