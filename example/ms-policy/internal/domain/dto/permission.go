package dto

import "github.com/google/uuid"

type PermissionProjection struct {
	Code string `bson:"code" json:"code"`
	Name string `bson:"name" json:"name"`
}

type PermissionCheck struct {
	UserID         uuid.UUID
	PermissionCode string
}

type PermissionCreate struct {
	RoleCode string
	Code     string
	Name     string
}

type PermissionCreateInput struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func NewPermissionProjection(code, name string) PermissionProjection {
	return PermissionProjection{
		Code: code,
		Name: name,
	}
}

func NewPermissionCheck(userID uuid.UUID, permissionCode string) PermissionCheck {
	return PermissionCheck{
		UserID:         userID,
		PermissionCode: permissionCode,
	}
}

func NewPermissionCreate(roleCode, code, name string) PermissionCreate {
	return PermissionCreate{
		RoleCode: roleCode,
		Code:     code,
		Name:     name,
	}
}

func NewPermissionCreateInput(code, name string) PermissionCreateInput {
	return PermissionCreateInput{
		Code: code,
		Name: name,
	}
}
