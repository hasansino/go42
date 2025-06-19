package domain

import (
	"encoding/json"
	"errors"
)

const DefaultFetchLimit = 10

var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
)

// ---

const (
	EventTypeCreated = iota
	EventTypeUpdated
	EventTypeDeleted
)

var EventTypes = map[int]string{
	EventTypeCreated: "fruit.created",
	EventTypeUpdated: "fruit.updated",
	EventTypeDeleted: "fruit.deleted",
}

// ---

type ExampleEvent struct {
	Type    int `json:"type"`
	Payload any `json:"payload"`
}

func (e *ExampleEvent) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ExampleEvent) Unmarshal(data []byte) error {
	return json.Unmarshal(data, e)
}
