package gateway

import (
	"context"
	"gateway/discovery"
	log "log/slog"

	pb "github.com/Abazin97/common/gen/go/order"
	"golang.org/x/exp/slog"
)

type Gateway struct {
	service discovery.Service
}

func NewGateway(service discovery.Service) *Gateway {
	return &Gateway{
		service: service,
	}
}

func (g *Gateway) UpdateOrder(ctx context.Context, orderID string, status string) error {
	conn, err := discovery.ServiceConnection(context.Background(), "order", g.service)
	if err != nil {
		log.Error("failed to dial server: ", err)
	}
	defer conn.Close()

	c := pb.NewOrderServiceClient(conn)

	_, err = c.UpdateOrder(ctx, &pb.UpdateOrderStatusRequest{
		OrderId: orderID,
		Status:  status,
	})
	if err != nil {
		log.Error("failed to update order", slog.String("error", err.Error()))
	}

	return err
}
