package entities

import (
	"errors"
	"time"
)

type UserActivityType int

const (
	Registration UserActivityType = iota + 1
	Authentication
	Verification
)

func (uat UserActivityType) String() string {
	return [...]string{"registration", "authentication", "verification"}[uat-1]
}

func (uat UserActivityType) Index() int {
	return int(uat)
}

type Activity struct {
	Type      UserActivityType
	Timestamp time.Time
	Device    string
}

func NewActivity(aType UserActivityType, payload map[string]interface{}) (*Activity, error) {
	a := &Activity{
		Type: aType,
	}
	return a.parse(payload)
}

func (a *Activity) parse(payload map[string]interface{}) (*Activity, error) {
	timestamp, ok := payload["timestamp"].(string)
	if !ok {
		return nil, errors.New("timestamp could not be parsed")
	}
	var err error
	a.Timestamp, err = time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return nil, err
	}
	device, ok := payload["device"].(string)
	if !ok {
		return nil, errors.New("device could not be parsed")
	}
	a.Device = device
	return a, nil
}
