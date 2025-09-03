package usecase

import (
	"api/internal/domain/dto"
	"api/internal/domain/entities"
	"api/internal/domain/helpers"
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
)

const (
	retryLimit = 5
	retryDelay = 100 * time.Millisecond
)

type PolicyCases interface {
	OnUpsert(ctx context.Context, id uuid.UUID, data dto.PolicyUpsert) error
}

type PolicyService struct {
	userRepository UserRepository
}

func (p *PolicyService) OnUpsert(ctx context.Context, id uuid.UUID, data dto.PolicyUpsert) error {
	userPolicy := dto.UserPolicy{
		ID:          id,
		Role:        data.Role,
		Permissions: data.Permissions,
	}
	seq := func(yield func(string) bool) {
		for _, user := range data.Users {
			if !yield(user.String()) {
				return
			}
		}
	}
	ids := slices.Collect(seq)
	if len(ids) == 0 {
		return nil
	}
	var users []*entities.User
	var err error
	helpers.Retry(
		func() bool {
			users, err = p.userRepository.Get(ctx, ids)
			userNotFound := false
			for _, userID := range ids {
				userNotFound = slices.IndexFunc(users, func(user *entities.User) bool {
					return user.ID().String() == userID
				}) == -1
				if userNotFound {
					break
				}
			}
			return err == nil && userNotFound
		},
		retryLimit,
		retryDelay,
	)

	if err != nil {
		return err
	}
	for _, user := range users {
		doc := user.Document()
		policies := doc.Policies
		if index := slices.IndexFunc(policies, func(x dto.UserPolicy) bool {
			return x.ID == id
		}); index == -1 {
			policies = append(policies, userPolicy)
		} else {
			policies[index] = userPolicy
		}
		err = p.userRepository.Upsert(ctx, entities.NewUser(doc.ID, doc.Info, policies))
		if err != nil {
			return err
		}
	}
	return nil
}

func NewPolicyService(userRepository UserRepository) *PolicyService {
	return &PolicyService{userRepository: userRepository}
}
