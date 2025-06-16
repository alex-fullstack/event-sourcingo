package entities

import (
	"errors"
)

type Credentials struct {
	Email        string
	PasswordHash string
	Confirmed    bool
}

func NewCredentials(payload map[string]interface{}) (*Credentials, error) {
	c := &Credentials{}
	return c.parse(payload)
}

func (c *Credentials) parse(payload map[string]interface{}) (*Credentials, error) {
	email, ok := payload["email"].(string)
	if !ok {
		return nil, errors.New("email could not be parsed")
	}
	c.Email = email
	passwordHash, ok := payload["password_hash"].(string)
	if !ok {
		return nil, errors.New("password hash could not be parsed")
	}
	c.PasswordHash = passwordHash
	return c, nil
}
