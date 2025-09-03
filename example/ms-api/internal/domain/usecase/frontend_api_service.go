package usecase

import (
	"api/internal/domain/dto"
	"context"
	"errors"

	"github.com/google/uuid"
)

type FrontendAPICases interface {
	Users(ctx context.Context, id uuid.UUID) ([]dto.UserDocument, error)
}

type FrontendAPIService struct {
	userRepository   UserRepository
	policyRepository PolicyRepository
}

func NewFrontendAPIService(
	userRepository UserRepository,
	policyRepository PolicyRepository,
) *FrontendAPIService {
	return &FrontendAPIService{userRepository: userRepository, policyRepository: policyRepository}
}

func (s *FrontendAPIService) Users(ctx context.Context, id uuid.UUID) ([]dto.UserDocument, error) {
	isAdmin := s.policyRepository.CheckRole(ctx, id, "admin")
	if !isAdmin {
		return []dto.UserDocument{}, errors.New("forbidden")
	}
	res := make([]dto.UserDocument, 0)
	users, err := s.userRepository.Get(ctx, make([]string, 0))
	if err != nil {
		return res, err
	}
	for _, user := range users {
		res = append(res, user.Document())
	}
	return res, nil
}
