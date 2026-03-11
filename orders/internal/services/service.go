package services

import (
	"context"
	"encoding/json"
	"fmt"
	"gateway/rabbitmq"
	"orders/internal/domain/models"
	"orders/internal/events"
	"orders/internal/gateway"
	"orders/internal/repository"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrdersService interface {
	CreateOrder(ctx context.Context, lotID string, customerID string, from time.Time, to time.Time, items []models.Item) (*models.Order, error)
	GetOrder(context.Context, string) (models.Order, error)
	UpdateOrder(ctx context.Context, id string, status string) error
}

type ordersService struct {
	repo    repository.Repository
	stock   gateway.StockGateway
	payment gateway.PaymentGateway
	channel *amqp.Channel
}

func NewOrdersService(repo repository.Repository, stockService gateway.StockGateway, paymentService gateway.PaymentGateway, channel *amqp.Channel) OrdersService {
	return &ordersService{repo: repo, stock: stockService, payment: paymentService, channel: channel}
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

	event := events.OrderCreatedEvent{
		OrderID:    orderID,
		LotID:      lotID,
		CustomerID: customerID,
		From:       from,
		To:         to,
		Items:      products,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = s.channel.PublishWithContext(ctx, rabbitmq.OrderExchange, rabbitmq.OrderCreatedEvent, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to publish order event")
	}

	//_, err = s.stock.Reserve(ctx, lotID, orderID, from, to)
	//if err != nil {
	//	return nil, fmt.Errorf("%s: %w", op, err)
	//}
	//
	//// todo: replace hardcode strings w/ special calculation method
	//pay, err := s.payment.CreatePayment(ctx, orderID, "2", "RUB")
	//if err != nil {
	//	// todo: releasing reservation in case of payment failure
	//	//s.stock.Release(ctx, reservationID)
	//	return nil, fmt.Errorf("%s: %w", op, err)
	//}
	//
	//err = s.repo.UpdatePaymentLink(ctx, orderID, pay.Confirmation.ConfirmationURL)
	//if err != nil {
	//	return nil, fmt.Errorf("%s: %w", op, err)
	//}

	o, err := s.repo.Get(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

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

func (s *ordersService) UpdateOrder(ctx context.Context, id string, status string) error {
	const op = "order.services.UpdateOrder"

	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
