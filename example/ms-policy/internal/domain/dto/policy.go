package dto

import (
	"errors"

	"github.com/google/uuid"
)

var ErrEmptyResult = errors.New("empty result")

type PolicyProjection struct {
	ID          uuid.UUID              `bson:"id"`
	Role        *RoleProjection        `bson:"role"`
	Permissions []PermissionProjection `bson:"permissions"`
	Users       []uuid.UUID            `bson:"users"`
}

func NewPolicyProjection(
	id uuid.UUID,
	role *RoleProjection,
	permissions []PermissionProjection,
	users []uuid.UUID,
) PolicyProjection {
	return PolicyProjection{
		ID:          id,
		Role:        role,
		Permissions: permissions,
		Users:       users,
	}
}
