package services

import (
	"context"
	"payments/internal/gateway"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, orderID string) error
}
type paymentService struct {
	gateway gateway.OrdersGateway
}

func NewPaymentService(ordersGateway gateway.OrdersGateway) PaymentService {
	return &paymentService{gateway: ordersGateway}
}

func (s *paymentService) CreatePayment(ctx context.Context, orderID string) error {

}
