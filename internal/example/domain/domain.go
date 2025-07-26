package domain

import (
	"errors"
)

const DefaultFetchLimit = 10

var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
)

// ---

const TopicNameFruitEvents = "fruit-events"

const (
	EventTypeCreated = "fruit.created"
	EventTypeUpdated = "fruit.updated"
	EventTypeDeleted = "fruit.deleted"
)
