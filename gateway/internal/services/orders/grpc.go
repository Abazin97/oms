package orders

import (
	"context"
	"gateway/internal/services"
	"time"

	pbo "github.com/Abazin97/common/gen/go/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ordersGateway struct {
	clientOrders pbo.OrderServiceClient
	conn         *grpc.ClientConn
}

func NewOrdersGateway(grpcAddr string) (services.OrdersGateway, error) {
	client, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	orderClient := pbo.NewOrderServiceClient(client)

	return &ordersGateway{
		clientOrders: orderClient,
	}, nil
}

func (g *ordersGateway) CreateOrder(ctx context.Context, cr *pbo.CreateOrderRequest) (*pbo.Order, error) {

	resp, err := g.clientOrders.CreateOrder(ctx, &pbo.CreateOrderRequest{
		CustomerId: cr.CustomerId,
		Id:         cr.Id,
		Items:      cr.Items,
	})
	if err != nil {
		return nil, err
	}

	// by Artem TOPSKIY

	return resp, nil
}

func (g *ordersGateway) GetOrder(ctx context.Context, id string) (*pbo.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := g.clientOrders.GetOrder(ctx, &pbo.GetOrderRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (g *ordersGateway) Close() error {
	return g.conn.Close()
}
