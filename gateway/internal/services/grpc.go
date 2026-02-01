package services

import (
	"context"
	"time"

	pb "github.com/Abazin97/common/gen/go/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type gateway struct {
	client pb.OrderServiceClient
	conn   *grpc.ClientConn
}

func NewGateway(grpcAddr string) (Gateway, error) {
	client, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	orderClient := pb.NewOrderServiceClient(client)

	return &gateway{
		client: orderClient,
	}, nil
}

func (g *gateway) CreateOrder(ctx context.Context, cr *pb.CreateOrderRequest) (*pb.Order, error) {
	//ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	//defer cancel()

	resp, err := g.client.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerId: cr.CustomerId,
		Items:      cr.Items,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (g *gateway) GetOrder(ctx context.Context, orderID int32) (*pb.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := g.client.GetOrder(ctx, &pb.GetOrderRequest{
		Id: string(orderID),
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (g *gateway) Close() error {
	return g.conn.Close()
}
