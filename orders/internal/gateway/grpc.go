package gateway

import (
	"context"
	"time"

	pbs "github.com/Abazin97/common/gen/go/stock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Gateway struct {
	clientStock pbs.StockServiceClient
	conn        *grpc.ClientConn
}

func NewStockGateway(grpcAddr string) (*Gateway, error) {
	client, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	stockClient := pbs.NewStockServiceClient(client)

	return &Gateway{
		clientStock: stockClient,
	}, nil
}

func (g *Gateway) Reserve(ctx context.Context, id string, orderID string, from time.Time, to time.Time) (*pbs.ReserveResponse, error) {
	return g.clientStock.Reserve(ctx, &pbs.ReserveRequest{
		LotId:   id,
		OrderId: orderID,
		To:      timestamppb.New(to),
		From:    timestamppb.New(from),
	})
}

func (g *Gateway) Close() error {
	return g.conn.Close()
}
