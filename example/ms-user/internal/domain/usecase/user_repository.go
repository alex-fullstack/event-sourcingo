package usecase

import (
	"context"
	"user/internal/domain/dto"

	"github.com/alex-fullstack/event-sourcingo/domain/usecases/repositories"
	"github.com/google/uuid"
)

type UserRepository interface {
	repositories.ProjectionSaver
	Get(ctx context.Context, id uuid.UUID) (dto.UserProjection, error)
	GetByEmail(ctx context.Context, email string) (dto.UserProjection, error)
}
