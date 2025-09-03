package usecase

import (
	"context"
	"policy/internal/domain/dto"

	"github.com/alex-fullstack/event-sourcingo/domain/usecases/repositories"
	"github.com/google/uuid"
)

type PolicyRepository interface {
	repositories.ProjectionSaver
	Get(ctx context.Context, id uuid.UUID) (*dto.PolicyProjection, error)
	GetByRoleCode(ctx context.Context, code string) (*dto.PolicyProjection, error)
	GetByUserID(ctx context.Context, id uuid.UUID) ([]dto.PolicyProjection, error)
}
