package usecase

import (
	"context"
	"errors"
	"time"
	"user/internal/domain/commands"
	"user/internal/domain/dto"
	"user/internal/domain/entities"
	"user/internal/domain/helpers"

	"github.com/alex-fullstack/event-sourcingo/domain/usecases/services"
	"github.com/go-ozzo/ozzo-validation/v4" //nolint:goimports // proper import
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
)

const (
	codeLength   = 8
	codeLifetime = 24 * time.Hour
)

type FrontendAPICases interface {
	Sign(ctx context.Context, data dto.UserSign) error
	Login(ctx context.Context, data dto.UserLogin) (dto.AuthResult, error)
	RequestConfirm(ctx context.Context, data dto.UserRequestConfirmation) error
	Confirm(ctx context.Context, data dto.ConfirmInput) error
}

type FrontendAPIService struct {
	repository UserRepository
	handler    services.CommandHandler
}

func NewFrontendAPIService(
	handler services.CommandHandler,
	repository UserRepository,
) *FrontendAPIService {
	return &FrontendAPIService{handler: handler, repository: repository}
}

func (s *FrontendAPIService) Sign(ctx context.Context, data dto.UserSign) error {
	if err := validation.ValidateStruct(&data.Credentials,
		validation.Field(&data.Credentials.Email, validation.Required, is.Email),
		validation.Field(
			&data.Credentials.Password,
			validation.Required,
			validation.Length(helpers.MinPasswordLen, helpers.MaxPasswordLen),
		),
	); err != nil {
		return err
	}

	return s.handler.Handle(
		ctx,
		commands.NewSignCommand(dto.NewCredentialsCreateInput(data.Credentials), data.Activity),
		entities.NewUser(uuid.New()),
	)
}

func (s *FrontendAPIService) Login(
	ctx context.Context,
	data dto.UserLogin,
) (dto.AuthResult, error) {
	user, err := s.repository.GetByEmail(ctx, data.Credentials.Email)
	if err != nil {
		return dto.AuthResult{}, err
	}
	err = helpers.CheckPass(data.Credentials.Password, user.Credentials.PasswordHash)
	if err != nil {
		return dto.AuthResult{}, err
	}
	err = s.handler.Handle(
		ctx,
		commands.NewLoginCommand(data.Activity),
		entities.NewUser(user.ID),
	)
	if err != nil {
		return dto.AuthResult{}, err
	}
	accessToken, refreshToken, err := helpers.Auth(user.ID)
	if err != nil {
		return dto.AuthResult{}, err
	}

	return dto.NewAuthResult(accessToken, refreshToken), nil
}

func (s *FrontendAPIService) RequestConfirm(
	ctx context.Context,
	data dto.UserRequestConfirmation,
) error {
	user, err := s.repository.Get(ctx, data.UserID)
	if err != nil {
		return err
	}
	code, err := helpers.RandomInt(codeLength)
	if err != nil {
		return err
	}
	return s.handler.Handle(
		ctx,
		commands.NewRequestConfirmCommand(
			dto.NewConfirmationInput(code, time.Now().Add(codeLifetime)),
			data.Activity,
			user.Confirmation == nil,
		),
		entities.NewUser(data.UserID),
	)
}

func (s *FrontendAPIService) Confirm(
	ctx context.Context,
	data dto.ConfirmInput,
) (confirmErr error) {
	user, err := s.repository.Get(ctx, data.UserID)
	if err != nil {
		return err
	}

	defer func() {
		handleErr := s.handler.Handle(
			ctx,
			commands.NewConfirmCommand(data.Activity, confirmErr != nil),
			entities.NewUser(data.UserID),
		)
		if confirmErr == nil {
			confirmErr = handleErr
		}
	}()
	confirmErr = validation.Date(time.RFC3339).
		Min(time.Now()).
		Validate(user.Confirmation.Expired.Format(time.RFC3339))
	if confirmErr != nil {
		return confirmErr
	}
	confirmErr = validation.ValidateStruct(
		&data,
		validation.Field(&data.Code, validation.By(func(value interface{}) error {
			if value != user.Confirmation.Code {
				return errors.New("wrong code")
			}
			return nil
		})),
	)
	return confirmErr
}
