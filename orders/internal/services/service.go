package services

import (
	"context"
	"fmt"
	"orders/internal/domain/models"
	"orders/internal/repository"

	pb "github.com/Abazin97/common/gen/go/order"
)

type OrdersService interface {
	CreateOrder(context.Context, string, []models.Item) (*models.Order, error)
	GetOrder(context.Context, string) (models.Order, error)
	UpdateOrder(context.Context, string, *pb.Order) (*pb.Order, error)
}

type ordersService struct {
	repo repository.Repository
}

func NewOrdersService(repo repository.Repository) OrdersService {
	return &ordersService{repo: repo}
}

func (s *ordersService) CreateOrder(ctx context.Context, customerID string, products []models.Item) (*models.Order, error) {
	const op = "order.services.CreateOrder"

	id, err := s.repo.Create(ctx, models.Order{
		CustomerId:  customerID,
		Items:       products,
		Status:      "pending",
		PaymentLink: "",
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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
	const op = "order.services.GetOrder"

	o, err := s.repo.Get(ctx, id)

	if err != nil {
		return models.Order{}, fmt.Errorf("%s: %w", op, err)
	}

	return o, nil
}

func (s *ordersService) UpdateOrder(ctx context.Context, id string, o *pb.Order) (*pb.Order, error) {
	const op = "order.services.UpdateOrder"

	err := s.repo.Update(ctx, id, models.Order{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return o, nil
}
