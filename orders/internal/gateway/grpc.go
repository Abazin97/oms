package gateway

import (
	"context"
	"gateway/discovery"
	"log"
	"orders/internal/domain/models"
	"time"

	pbp "github.com/Abazin97/common/gen/go/payment"
	pbs "github.com/Abazin97/common/gen/go/stock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type stockGateway struct {
	service discovery.Service
}

type paymentGateway struct {
	service discovery.Service
}

func NewStockGateway(service discovery.Service) StockGateway {
	return &stockGateway{
		service: service,
	}
}

func NewPaymentGateway(service discovery.Service) PaymentGateway {
	return &paymentGateway{
		service: service,
	}
}

func (g *stockGateway) Reserve(ctx context.Context, lotID string, orderID string, from time.Time, to time.Time) (*pbs.ReserveResponse, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "stock", g.service)
	if err != nil {
		log.Fatal("failed to dial server: ", err)
	}
	defer conn.Close()

	c := pbs.NewStockServiceClient(conn)
	return c.Reserve(ctx, &pbs.ReserveRequest{
		Id:      lotID,
		OrderId: orderID,
		To:      timestamppb.New(to),
		From:    timestamppb.New(from),
	})
}

func (g *paymentGateway) CreatePayment(ctx context.Context, orderID string, amount string, currency string) (*models.YouKassaResponse, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "payment", g.service)
	if err != nil {
		log.Fatal("failed to dial server: ", err)
	}
	defer conn.Close()

	c := pbp.NewPaymentServiceClient(conn)
	resp, err := c.CreatePayment(ctx, &pbp.CreatePaymentRequest{
		OrderId:  orderID,
		Amount:   amount,
		Currency: currency,
	})
	if err != nil {
		return nil, err
	}

	return &models.YouKassaResponse{
		ID: resp.PaymentId,
		Confirmation: struct {
			ConfirmationURL string `json:"confirmation_url"`
		}{ConfirmationURL: resp.PaymentUrl},
		Status: resp.Status,
		Amount: struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		}{Value: amount, Currency: currency},
	}, nil
}
