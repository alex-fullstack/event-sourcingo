package usecase

import (
	"api/internal/domain/entities"
	"context"
)

type UserRepository interface {
	Get(ctx context.Context, ids []string) ([]*entities.User, error)
	Upsert(ctx context.Context, user *entities.User) error
}
