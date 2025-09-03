package usecase

import (
	"context"
	"policy/internal/domain/commands"
	"policy/internal/domain/entities"

	"github.com/alex-fullstack/event-sourcingo/domain/usecases/services"
	"github.com/google/uuid"
)

type MSUserCases interface {
	OnUserCreate(ctx context.Context, userID uuid.UUID) error
}

type UserService struct {
	repository PolicyRepository
	handler    services.CommandHandler
}

func NewUserService(handler services.CommandHandler, repository PolicyRepository) *UserService {
	return &UserService{handler: handler, repository: repository}
}

func (s *UserService) OnUserCreate(ctx context.Context, userID uuid.UUID) error {
	policy, err := s.repository.GetByRoleCode(ctx, entities.UserRole.String())
	if err != nil {
		return err
	}

	return s.handler.Handle(
		ctx,
		commands.NewUserAssignCommand(userID),
		entities.NewPolicy(policy.ID),
	)
}
