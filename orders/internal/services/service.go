package services

import (
	"context"
	"fmt"
	log "log/slog"
	"orders/internal/domain/models"
	"orders/internal/gateway"
	"orders/internal/repository"
	"time"
)

type OrdersService interface {
	CreateOrder(context.Context, string, string, time.Time, time.Time, []models.Item) (*models.Order, error)
	GetOrder(context.Context, string) (models.Order, error)
	UpdateOrder(context.Context, string, string) (*models.Order, error)
}

type ordersService struct {
	repo    repository.Repository
	stock   gateway.StockGateway
	payment gateway.PaymentGateway
}

func NewOrdersService(repo repository.Repository, stockService gateway.StockGateway, paymentService gateway.PaymentGateway) OrdersService {
	return &ordersService{repo: repo, stock: stockService, payment: paymentService}
}

func (s *ordersService) CreateOrder(ctx context.Context, lotID string, customerID string, from time.Time, to time.Time, products []models.Item) (*models.Order, error) {
	const op = "order.services.CreateOrder"

	orderID, err := s.repo.Create(ctx, models.Order{
		CustomerId:  customerID,
		Items:       products,
		Status:      "pending",
		PaymentLink: "",
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("orderID: ", orderID,
		"customerID: ", customerID,
		"products: ", products,
		"fromTime: ", from,
		"toTime: ", to)

	_, err = s.stock.Reserve(ctx, lotID, orderID, from, to)

	pay, err := s.payment.CreatePayment(ctx, orderID, "2", "RUB")
	if err != nil {
		// todo: releasing reservation in case of payment failure
		//s.stock.Release(ctx, reservationID)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = s.repo.Update(ctx, orderID, pay.Confirmation.ConfirmationURL)

	o, err := s.repo.Get(ctx, orderID)

	return &o, nil
}

func (s *ordersService) GetOrder(ctx context.Context, id string) (models.Order, error) {
	const op = "order.services.GetOrder"

	o, err := s.repo.Get(ctx, id)

	if err != nil {
		return models.Order{}, fmt.Errorf("%s: %w", op, err)
	}

	return o, nil
}

func (s *ordersService) UpdateOrder(ctx context.Context, id string, status string) (*models.Order, error) {
	const op = "order.services.UpdateOrder"

	err := s.repo.Update(ctx, id, status)

	o, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &o, nil
}
