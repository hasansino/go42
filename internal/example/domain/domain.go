package domain

import (
	"encoding/json"
	"errors"
)

const (
	DefaultFetchLimit = 10

	EventTypeCreated = "fruit.created"
	EventTypeUpdated = "fruit.updated"
	EventTypeDeleted = "fruit.deleted"
)

var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
)

type CreateFruitRequest struct {
	Name string `json:"name" v:"required,min=3,max=20"`
}

type UpdateFruitRequest struct {
	Name string `json:"name" v:"required,min=3,max=20"`
}

// ---

type ExampleEvent struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func (e *ExampleEvent) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ExampleEvent) Unmarshal(data []byte) error {
	return json.Unmarshal(data, e)
}
