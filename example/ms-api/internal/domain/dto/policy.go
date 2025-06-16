package dto

import (
	"errors"

	"github.com/google/uuid"
)

type PolicyEventType string

const (
	RoleCreated       PolicyEventType = "role_created"
	PermissionCreated PolicyEventType = "permission_created"
	UserAssigned      PolicyEventType = "user_assigned"
)

type Role struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Permission struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type PolicyUpsert struct {
	Role        *Role        `json:"role"`
	Permissions []Permission `json:"permissions"`
	Users       []uuid.UUID  `json:"users"`
}

func NewPolicyUpsert(payload map[string]interface{}) (PolicyUpsert, error) {
	pu := &PolicyUpsert{Permissions: make([]Permission, 0), Users: make([]uuid.UUID, 0)}
	return pu.parse(payload)
}

func (pu *PolicyUpsert) parse(payload map[string]interface{}) (PolicyUpsert, error) {
	rawRole := payload["role"]
	if rawRole != nil {
		role, ok := rawRole.(map[string]interface{})
		if !ok {
			return PolicyUpsert{}, errors.New("role could not be parsed")
		}
		roleCode, ok := role["code"].(string)
		if !ok {
			return PolicyUpsert{}, errors.New("role code could not be parsed")
		}
		roleName, ok := role["name"].(string)
		if !ok {
			return PolicyUpsert{}, errors.New("role name could not be parsed")
		}
		pu.Role = &Role{
			Code: roleCode,
			Name: roleName,
		}
	}
	var err error
	pu.Permissions, err = collectPermissions(payload)
	if err != nil {
		return PolicyUpsert{}, err
	}
	pu.Users, err = collectUsers(payload)
	if err != nil {
		return PolicyUpsert{}, err
	}
	return *pu, nil
}

func collectPermissions(payload map[string]interface{}) ([]Permission, error) {
	var ps []Permission
	permissions, ok := payload["permissions"].([]interface{})
	if !ok {
		return nil, errors.New("permissions could not be parsed")
	}
	for _, permission := range permissions {
		perm, permOk := permission.(map[string]interface{})
		if !permOk {
			return nil, errors.New("permission could not be parsed")
		}
		permissionCode, codeOk := perm["code"].(string)
		if !codeOk {
			return nil, errors.New("permission code could not be parsed")
		}
		permissionName, nameOk := perm["name"].(string)
		if !nameOk {
			return nil, errors.New("permission name could not be parsed")
		}
		ps = append(ps, Permission{
			Code: permissionCode,
			Name: permissionName,
		})
	}
	return ps, nil
}

func collectUsers(payload map[string]interface{}) ([]uuid.UUID, error) {
	var uids []uuid.UUID
	users, ok := payload["users"].([]interface{})
	if !ok {
		return nil, errors.New("users could not be parsed")
	}
	for _, user := range users {
		id, userOk := user.(string)
		if !userOk {
			return nil, errors.New("user id could not be parsed")
		}
		userID, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		uids = append(uids, userID)
	}
	return uids, nil
}
