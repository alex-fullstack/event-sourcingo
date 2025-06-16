package dto

import (
	"time"
)

type ActivityInput struct {
	Timestamp time.Time `json:"timestamp"`
	Device    string    `json:"device"`
}

type ActivityRecord struct {
	Type      string    `bson:"type"`
	Timestamp time.Time `bson:"timestamp"`
	Device    string    `bson:"device"`
}

func NewActivityRecord(timestamp time.Time, device string, aType string) ActivityRecord {
	return ActivityRecord{Timestamp: timestamp, Device: device, Type: aType}
}

func NewActivityInput(timestamp time.Time, device string) ActivityInput {
	return ActivityInput{Timestamp: timestamp, Device: device}
}
