package cli

import (
	"errors"
	"policy/internal/domain/dto"
)

const (
	twoArgs   = 2
	threeArgs = 3
)

type Converter interface {
	ConvertRoleCreate(args ...string) (dtoRole dto.RoleCreate, err error)
	ConvertPermissionCreate(args ...string) (dtoPermission dto.PermissionCreate, err error)
	ConvertUserAssign(args ...string) (dtoUserAssign dto.UserAssign, err error)
}
type converter struct{}

func NewConverter() Converter {
	return &converter{}
}

func (c *converter) ConvertRoleCreate(args ...string) (dto.RoleCreate, error) {
	if len(args) < twoArgs {
		return dto.RoleCreate{}, errors.New("not enough arguments")
	}
	return dto.NewRoleCreate(args[0], args[1]), nil
}

func (c *converter) ConvertPermissionCreate(args ...string) (dto.PermissionCreate, error) {
	if len(args) < threeArgs {
		return dto.PermissionCreate{}, errors.New("not enough arguments")
	}
	return dto.NewPermissionCreate(args[0], args[1], args[2]), nil
}

func (c *converter) ConvertUserAssign(args ...string) (dto.UserAssign, error) {
	if len(args) < twoArgs {
		return dto.UserAssign{}, errors.New("not enough arguments")
	}
	return dto.NewUserAssign(args[0], args[1]), nil
}
