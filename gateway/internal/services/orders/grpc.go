package orders

import (
	"context"
	discovery "gateway/discovery"
	"gateway/internal/services"
	"log"

	pbo "github.com/Abazin97/common/gen/go/order"
)

type ordersGateway struct {
	service discovery.Service
}

func NewOrdersGateway(service discovery.Service) services.OrdersGateway {
	return &ordersGateway{
		service: service,
	}
}

func (g *ordersGateway) CreateOrder(ctx context.Context, cr *pbo.CreateOrderRequest) (*pbo.Order, error) {
	conn, err := discovery.ServiceConnection(context.Background(), "orders", g.service)
	if err != nil {
		log.Fatal("failed to dial server: ", err)
	}
	defer conn.Close()

	c := pbo.NewOrderServiceClient(conn)

	resp, err := c.CreateOrder(ctx, &pbo.CreateOrderRequest{
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
	conn, err := discovery.ServiceConnection(context.Background(), "orders", g.service)
	if err != nil {
		log.Fatal("failed to dial server: ", err)
	}
	defer conn.Close()

	c := pbo.NewOrderServiceClient(conn)
	resp, err := c.GetOrder(ctx, &pbo.GetOrderRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
