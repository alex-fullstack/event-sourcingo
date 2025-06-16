package dto

import "github.com/google/uuid"

type RoleProjection struct {
	Code string `bson:"code" json:"code"`
	Name string `bson:"name" json:"name"`
}

type RoleCreate struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type RoleCheck struct {
	UserID   uuid.UUID
	RoleCode string
}

func NewRoleProjection(code, name string) *RoleProjection {
	return &RoleProjection{Code: code, Name: name}
}

func NewRoleCreate(code, name string) RoleCreate {
	return RoleCreate{Code: code, Name: name}
}

func NewRoleCheck(userID uuid.UUID, roleCode string) RoleCheck {
	return RoleCheck{
		UserID:   userID,
		RoleCode: roleCode,
	}
}
