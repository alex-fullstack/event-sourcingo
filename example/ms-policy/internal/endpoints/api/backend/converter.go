package backend

import (
	"policy/internal/domain/dto"
	v1 "policy/internal/endpoints/api/backend/generated/v1"

	"github.com/google/uuid"
)

type Converter interface {
	ConvertRoleCheck(*v1.RoleCheckRequest) (dto.RoleCheck, error)
	ConvertPermissionCheck(*v1.PermissionCheckRequest) (dto.PermissionCheck, error)
}

type converter struct{}

func NewConverter() Converter {
	return converter{}
}

func (c converter) ConvertRoleCheck(req *v1.RoleCheckRequest) (dto.RoleCheck, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return dto.RoleCheck{}, err
	}
	return dto.NewRoleCheck(userID, req.GetRoleCode()), nil
}

func (c converter) ConvertPermissionCheck(
	req *v1.PermissionCheckRequest,
) (dto.PermissionCheck, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return dto.PermissionCheck{}, err
	}
	return dto.NewPermissionCheck(userID, req.GetPermissionCode()), nil
}
