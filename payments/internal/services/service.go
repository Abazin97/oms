package services

import (
	"context"
	"fmt"
	"payments/internal/domain/models"
	"payments/internal/gateway"
	"payments/internal/yookassa"
)

const checkoutURL = "https://abazincloud.ddns.net"

type PaymentService interface {
	CreatePayment(ctx context.Context, orderID string, amount string, currency string) (*models.YouKassaResponse, error)
	HandleYouKassaWebHook(ctx context.Context, n models.YouKassaNotification) error
	//UpdatePayment(ctx context.Context, orderID string, paymentLink string) error
}
type paymentService struct {
	gateway  gateway.OrdersGateway
	yooKassa *yookassa.Client
}

func NewPaymentService(ordersGateway gateway.OrdersGateway, yooKassa *yookassa.Client) PaymentService {
	return &paymentService{gateway: ordersGateway, yooKassa: yooKassa}
}

func (s *paymentService) CreatePayment(ctx context.Context, orderID string, amount string, currency string) (*models.YouKassaResponse, error) {
	const op = "payments.services.CreatePayment"

	req := models.YouKassaRequest{
		Amount: models.Amount{
			Currency: currency,
			Value:    amount,
		},
		Confirmation: models.Confirmation{
			Type:      "redirect",
			ReturnURL: checkoutURL,
		},
		Capture:     true,
		Description: "parking payment",
	}

	return s.yooKassa.CreatePayment(ctx, orderID, req)
}

func (s *paymentService) HandleYouKassaWebHook(ctx context.Context, n models.YouKassaNotification) error {
	const op = "payment.services.HandleWebhook"

	if n.Event != "payment.succeeded" {
		return nil
	}

	//err := s.repo.UpdateStatus(ctx, n.Object.ID, "succeeded")
	//if err != nil {
	//	return fmt.Errorf("%s: %w", op, err)
	//}

	err := s.gateway.UpdateOrder(ctx, n.Object.Metadata.OrderID, "paid")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

//func (s *paymentService) UpdatePayment(ctx context.Context, orderID string, paymentLink string) error {
//
//}
