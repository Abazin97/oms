package stock

import (
	"context"
	"gateway/internal/services"

	pbs "github.com/Abazin97/common/gen/go/stock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type stockGateway struct {
	clientStock pbs.StockServiceClient
	conn        *grpc.ClientConn
}

func NewStockGateway(grpcAddr string) (services.StockGateway, error) {
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

func (g *stockGateway) GetStock(ctx context.Context, cr *pbs.GetAvailabilityRequest) (*pbs.GetAvailabilityResponse, error) {
	return g.clientStock.GetAvailability(ctx, cr)
}

func (g *stockGateway) Close() error {
	return g.conn.Close()
}
