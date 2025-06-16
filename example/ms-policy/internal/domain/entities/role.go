package entities

import (
	"errors"
)

type RoleCodeType string

const (
	UserRole      RoleCodeType = "user"
	AdminRole     RoleCodeType = "admin"
	UserRoleName  string       = "Пользователь"
	AdminRoleName string       = "Администратор"
)

func (role RoleCodeType) String() string {
	return string(role)
}

type Role struct {
	Code RoleCodeType
	Name string
}

func NewRole(payload map[string]interface{}) (*Role, error) {
	r := &Role{}
	return r.parse(payload)
}

func (r *Role) parse(payload map[string]interface{}) (*Role, error) {
	code, ok := payload["code"].(string)
	if !ok {
		return nil, errors.New("code could not be parsed")
	}
	r.Code = RoleCodeType(code)
	name, ok := payload["name"].(string)
	if !ok {
		return nil, errors.New("name could not be parsed")
	}
	r.Name = name
	return r, nil
}
