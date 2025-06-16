package usecase

import (
	"context"
	"errors"
	"policy/internal/domain/dto"
	"slices"

	"gitverse.ru/aleksandr-bebyakov/event-sourcingo/domain/usecases/services"
)

type BackendAPICases interface {
	CheckRole(ctx context.Context, data dto.RoleCheck) error
	CheckPermission(ctx context.Context, data dto.PermissionCheck) error
}

type BackendAPIService struct {
	repository PolicyRepository
	handler    services.CommandHandler
}

func NewBackendAPIService(handler services.CommandHandler, repository PolicyRepository) *BackendAPIService {
	return &BackendAPIService{handler: handler, repository: repository}
}

func (s *BackendAPIService) CheckRole(ctx context.Context, data dto.RoleCheck) error {
	policies, err := s.repository.GetByUserID(ctx, data.UserID)
	if err != nil {
		return err
	}
	if len(policies) == 0 {
		return errors.New("user not found")
	}
	for _, policy := range policies {
		if policy.Role != nil && policy.Role.Code == data.RoleCode {
			return nil
		}
	}

	return errors.New("role code not found")
}

func (s *BackendAPIService) CheckPermission(ctx context.Context, data dto.PermissionCheck) error {
	policies, err := s.repository.GetByUserID(ctx, data.UserID)
	if err != nil {
		return err
	}
	if len(policies) == 0 {
		return errors.New("user not found")
	}
	for _, policy := range policies {
		permCodes := slices.Collect(func(yield func(string) bool) {
			for _, perm := range policy.Permissions {
				if !yield(perm.Code) {
					return
				}
			}
		})
		if slices.Index(permCodes, data.PermissionCode) != -1 {
			return nil
		}
	}

	return errors.New("permission code not found")
}
