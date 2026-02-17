package gateway

import (
	"context"
	"orders/internal/domain/models"
	"time"

	pbp "github.com/Abazin97/common/gen/go/payment"
	pbs "github.com/Abazin97/common/gen/go/stock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type stockGateway struct {
	clientStock pbs.StockServiceClient
	conn        *grpc.ClientConn
}

type paymentGateway struct {
	client pbp.PaymentServiceClient
	conn   *grpc.ClientConn
}

func NewStockGateway(grpcAddr string) (StockGateway, error) {
	client, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	stockClient := pbs.NewStockServiceClient(client)

	return &stockGateway{
		clientStock: stockClient,
		conn:        client,
	}, nil
}

func NewPaymentGateway(grpcAddr string) (PaymentGateway, error) {
	conn, err := grpc.NewClient(grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &paymentGateway{
		client: pbp.NewPaymentServiceClient(conn),
		conn:   conn,
	}, nil
}

func (g *stockGateway) Reserve(ctx context.Context, lotID string, orderID string, from time.Time, to time.Time) (*pbs.ReserveResponse, error) {
	return g.clientStock.Reserve(ctx, &pbs.ReserveRequest{
		Id:      lotID,
		OrderId: orderID,
		To:      timestamppb.New(to),
		From:    timestamppb.New(from),
	})
}

func (g *stockGateway) Close() error {
	return g.conn.Close()
}

func (g *paymentGateway) CreatePayment(ctx context.Context, orderID string, amount string, currency string) (*models.YouKassaResponse, error) {
	resp, err := g.client.CreatePayment(ctx, &pbp.CreatePaymentRequest{
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
