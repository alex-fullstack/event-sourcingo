package usecase

import (
	"context"
	"policy/internal/domain/dto"

	"github.com/google/uuid"
	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/usecases/repositories"
)

type PolicyRepository interface {
	repositories.ProjectionSaver
	Get(ctx context.Context, id uuid.UUID) (*dto.PolicyProjection, error)
	GetByRoleCode(ctx context.Context, code string) (*dto.PolicyProjection, error)
	GetByUserID(ctx context.Context, id uuid.UUID) ([]dto.PolicyProjection, error)
}
