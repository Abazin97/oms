package services

import (
	"context"
	"orders/internal/domain/models"
	"orders/internal/repository"

	pb "github.com/Abazin97/common/gen/go/order"
)

type OrdersService interface {
	CreateOrder(context.Context, string, []models.Item) (*models.Order, error)
	GetOrder(context.Context, string) (models.Order, error)
	UpdateOrder(context.Context) (*pb.Order, error)
}

type ordersService struct {
	repo repository.Repository
}

func NewOrdersService(repo repository.Repository) OrdersService {
	return &ordersService{repo: repo}
}

func (s *ordersService) CreateOrder(ctx context.Context, customerID string, products []models.Item) (*models.Order, error) {
	id, err := s.repo.Create(ctx, models.Order{
		CustomerId:  customerID,
		Items:       products,
		Status:      "pending",
		PaymentLink: "",
	})
	if err != nil {
		return nil, err
	}

	o := &models.Order{
		CustomerId: customerID,
		Items:      products,
		Status:     "pending",
		Id:         id,
	}

	return o, nil
}

func (s *ordersService) GetOrder(ctx context.Context, id string) (models.Order, error) {
	o, err := s.repo.Get(ctx, id)

	if err != nil {
		return models.Order{}, err
	}

	return o, nil
}

func (s *ordersService) UpdateOrder(ctx context.Context) (*pb.Order, error) {
	err := s.repo.Update(ctx, models.Order{})
	if err != nil {
		return &pb.Order{}, err
	}

	return &pb.Order{}, nil
}
