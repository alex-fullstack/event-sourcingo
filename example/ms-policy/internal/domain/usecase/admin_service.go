package usecase

import (
	"context"
	"errors"
	"policy/internal/domain/commands"
	"policy/internal/domain/dto"
	"policy/internal/domain/entities"
	"regexp"
	"slices"

	"github.com/alex-fullstack/event-sourcingo/domain/usecases/services"
	"github.com/asaskevich/govalidator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
)

type AdminCases interface {
	CreateRole(ctx context.Context, data dto.RoleCreate) error
	CreatePermission(ctx context.Context, data dto.PermissionCreate) error
	AssignUser(ctx context.Context, data dto.UserAssign) error
}

type AdminService struct {
	repository     PolicyRepository
	userRepository UserRepository
	handler        services.CommandHandler
}

func NewAdminService(
	handler services.CommandHandler,
	repository PolicyRepository,
	userRepository UserRepository,
) *AdminService {
	return &AdminService{handler: handler, repository: repository, userRepository: userRepository}
}

func (s *AdminService) CreatePermission(ctx context.Context, data dto.PermissionCreate) error {
	if err := validation.ValidateStruct(&data,
		validation.Field(&data.RoleCode, validation.Required, is.LowerCase, is.Alpha),
		validation.Field(&data.Code, validation.Required, is.LowerCase, is.Alpha),
		validation.Field(
			&data.Name,
			validation.Required,
			validation.NewStringRuleWithError(
				func(str string) bool {
					if govalidator.IsNull(str) {
						return true
					}
					return regexp.MustCompile("^[\u0401\u0451\u0410-\u044f]+$").MatchString(str)
				},
				validation.NewError("validation_is_rus", "must contain Russian letters only")),
		),
	); err != nil {
		return err
	}

	policy, err := s.repository.GetByRoleCode(ctx, data.RoleCode)
	if err != nil {
		return err
	}

	return s.handler.Handle(
		ctx,
		commands.NewCreatePermissionCommand(dto.NewPermissionCreateInput(data.Code, data.Name)),
		entities.NewPolicy(policy.ID),
	)
}

func (s *AdminService) CreateRole(ctx context.Context, data dto.RoleCreate) error {
	if err := validation.ValidateStruct(&data,
		validation.Field(&data.Code, validation.Required, is.LowerCase, is.Alpha),
		validation.Field(
			&data.Name,
			validation.Required,
			validation.NewStringRuleWithError(
				func(str string) bool {
					if govalidator.IsNull(str) {
						return true
					}
					return regexp.MustCompile("^[\u0401\u0451\u0410-\u044f]+$").MatchString(str)
				},
				validation.NewError("validation_is_rus", "must contain Russian letters only")),
		),
	); err != nil {
		return err
	}

	return s.handler.Handle(
		ctx,
		commands.NewCreateRoleCommand(data),
		entities.NewPolicy(uuid.New()),
	)
}

func (s *AdminService) AssignUser(ctx context.Context, data dto.UserAssign) error {
	id, err := s.userRepository.GetUserIDByEmail(ctx, data.Email)
	if err != nil {
		return err
	}
	policy, err := s.repository.GetByRoleCode(ctx, data.Role)
	if err != nil {
		return err
	}
	err = validation.Validate(policy.Users, validation.By(func(value interface{}) error {
		ids, _ := value.([]uuid.UUID)
		if slices.Contains(ids, id) {
			return errors.New("user already assigned")
		}
		return nil
	}))
	if err != nil {
		return err
	}
	return s.handler.Handle(
		ctx,
		commands.NewUserAssignCommand(id),
		entities.NewPolicy(policy.ID),
	)
}
