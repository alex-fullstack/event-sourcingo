package entities

import (
	"api/internal/domain/dto"
	"slices"

	"github.com/google/uuid"
)

type Role struct {
	Code string
	Name string
}

type Permission struct {
	Code string
	Name string
}

type Policy struct {
	id          uuid.UUID
	role        *Role
	permissions []Permission
}

func NewPolicy(policy dto.UserPolicy) *Policy {
	var role *Role
	if policy.Role != nil {
		role = &Role{
			Code: policy.Role.Code,
			Name: policy.Role.Name,
		}
	}
	p := &Policy{
		id:          policy.ID,
		role:        role,
		permissions: make([]Permission, 0),
	}
	if len(policy.Permissions) > 0 {
		p.permissions = slices.Collect(func(yield func(permission Permission) bool) {
			for _, perm := range policy.Permissions {
				if !yield(
					Permission{
						Code: perm.Code,
						Name: perm.Name,
					}) {
					return
				}
			}
		})
	}
	return p
}

func (p *Policy) Document() dto.UserPolicy {
	var role *dto.Role
	if p.role != nil {
		role = &dto.Role{
			Code: p.role.Code,
			Name: p.role.Name,
		}
	}
	up := dto.UserPolicy{
		ID:          p.id,
		Role:        role,
		Permissions: make([]dto.Permission, 0),
	}
	if len(p.permissions) > 0 {
		up.Permissions = slices.Collect(func(yield func(permission dto.Permission) bool) {
			for _, perm := range p.permissions {
				if !yield(dto.Permission{
					Code: perm.Code,
					Name: perm.Name,
				}) {
					return
				}
			}
		})
	}
	return up
}
