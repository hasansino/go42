package domain

import "errors"

const (
	DefaultFetchLimit = 10
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
