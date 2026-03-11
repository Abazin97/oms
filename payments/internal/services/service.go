package services

import (
	"context"
	"encoding/json"
	"gateway/rabbitmq"
	"payments/internal/domain/models"
	"payments/internal/events"
	"payments/internal/gateway"
	"payments/internal/yookassa"

	amqp "github.com/rabbitmq/amqp091-go"
)

const checkoutURL = "https://abazincloud.ddns.net"

type PaymentService interface {
	CreatePayment(ctx context.Context, orderID string, amount string, currency string) (*models.YouKassaResponse, error)
	HandleYouKassaWebHook(ctx context.Context, n models.YouKassaNotification) error
	//UpdatePayment(ctx context.Context, orderID string, paymentLink string) error
}
type paymentService struct {
	gateway gateway.OrdersGateway

	yooKassa *yookassa.Client
	channel  *amqp.Channel
}

func NewPaymentService(ordersGateway gateway.OrdersGateway, yooKassa *yookassa.Client, channel *amqp.Channel) PaymentService {
	return &paymentService{gateway: ordersGateway, yooKassa: yooKassa, channel: channel}
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
		Metadata: map[string]string{
			"orderId": orderID,
		},
	}

	pay, err := s.yooKassa.CreatePayment(ctx, orderID, req)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(pay)
	if err != nil {
		return nil, err
	}

	err = s.channel.PublishWithContext(ctx, rabbitmq.OrderExchange, rabbitmq.PaymentCreatedEvent, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	})

	return pay, nil
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

	event := events.OrderPaidEvent{
		OrderID:  n.Object.Metadata["orderId"],
		Amount:   n.Object.Amount.Value,
		Currency: n.Object.Amount.Currency,
		Status:   "paid",
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.channel.PublishWithContext(ctx, rabbitmq.OrderExchange, rabbitmq.OrderPaidEvent, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	})
	//err := s.gateway.UpdateOrder(ctx, n.Object.Metadata["orderId"], "paid")
	//if err != nil {
	//	return fmt.Errorf("%s: %w", op, err)
	//}
}

//func (s *paymentService) UpdatePayment(ctx context.Context, orderID string, paymentLink string) error {
//
//}
