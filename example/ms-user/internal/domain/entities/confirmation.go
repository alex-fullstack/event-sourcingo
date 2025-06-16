package entities

import (
	"errors"
	"time"
)

type Confirmation struct {
	Code    string
	Expired time.Time
}

func NewConfirmation(payload map[string]interface{}) (*Confirmation, error) {
	c := &Confirmation{}
	return c.parse(payload)
}

func (c *Confirmation) parse(payload map[string]interface{}) (*Confirmation, error) {
	code, ok := payload["code"].(string)
	if !ok {
		return nil, errors.New("code could not be parsed")
	}
	c.Code = code
	expired, ok := payload["expired"].(string)
	if !ok {
		return nil, errors.New("expired could not be parsed")
	}
	var err error
	c.Expired, err = time.Parse(time.RFC3339, expired)
	if err != nil {
		return nil, err
	}
	return c, nil
}
