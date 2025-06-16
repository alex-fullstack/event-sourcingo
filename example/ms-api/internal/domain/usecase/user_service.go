package usecase

import (
	"api/internal/domain/dto"
	"api/internal/domain/entities"
	"context"

	"github.com/google/uuid"
)

type UserCases interface {
	OnUpsert(ctx context.Context, id uuid.UUID, data dto.UserUpsert) error
}

type UserService struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository: repository}
}

func (u *UserService) OnUpsert(ctx context.Context, id uuid.UUID, data dto.UserUpsert) error {
	docs, err := u.repository.Get(ctx, []string{id.String()})
	if err != nil {
		return err
	}
	policies := make([]dto.UserPolicy, 0)
	if len(docs) == 1 {
		policies = docs[0].Document().Policies
	}
	return u.repository.Upsert(ctx, entities.NewUser(id, data, policies))
}
