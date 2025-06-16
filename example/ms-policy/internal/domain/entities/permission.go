package entities

import "errors"

type Permission struct {
	Code string
	Name string
}

func NewPermission(payload map[string]interface{}) (*Permission, error) {
	p := &Permission{}
	return p.parse(payload)
}

func (p *Permission) parse(payload map[string]interface{}) (*Permission, error) {
	code, ok := payload["code"].(string)
	if !ok {
		return nil, errors.New("code could not be parsed")
	}
	p.Code = code
	name, ok := payload["name"].(string)
	if !ok {
		return nil, errors.New("name could not be parsed")
	}
	p.Name = name
	return p, nil
}
