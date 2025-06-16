package usecase

import (
	"context"
)

type BackendAPICases interface {
	GetUserIDByEmail(ctx context.Context, email string) (string, error)
}

type BackendAPIService struct {
	repository UserRepository
}

func NewBackendAPIService(repository UserRepository) *BackendAPIService {
	return &BackendAPIService{repository: repository}
}

func (s *BackendAPIService) GetUserIDByEmail(ctx context.Context, email string) (string, error) {
	user, err := s.repository.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	return user.ID.String(), err
}
