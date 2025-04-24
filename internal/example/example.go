package example

import (
	"golang.org/x/net/context"

	"github.com/hasansino/goapp/internal/example/models"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Fruits(_ context.Context) ([]models.Fruit, error) {
	return nil, nil
}

func (s *Service) FruitByID(_ context.Context, _ int) (*models.Fruit, error) {
	return nil, nil
}
