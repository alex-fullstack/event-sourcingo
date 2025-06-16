package usecase

import (
	"context"
	"user/internal/domain/dto"

	"github.com/google/uuid"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/usecases/repositories"
)

type UserRepository interface {
	repositories.ProjectionSaver
	Get(ctx context.Context, id uuid.UUID) (dto.UserProjection, error)
	GetByEmail(ctx context.Context, email string) (dto.UserProjection, error)
}
