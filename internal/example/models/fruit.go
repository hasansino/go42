package models

import (
	"gorm.io/gorm"
)

type Fruit struct {
	gorm.Model
	Name string `json:"name"`
}

func (m *Fruit) TableName() string {
	return "example_fruits"
}
