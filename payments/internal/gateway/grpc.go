package gateway

import (
	"context"

	pb "github.com/Abazin97/common/gen/go/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	clientOrder pb.OrderServiceClient
	conn        *grpc.ClientConn
}

func NewGateway(grpcAddr string) (*Gateway, error) {
	client, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	orderClient := pb.NewOrderServiceClient(client)

	return &Gateway{
		clientOrder: orderClient,
		conn:        client,
	}, nil
}

func (g *Gateway) UpdateOrder(ctx context.Context, orderID string, paymentLink string) error {
	_, err := g.clientOrder.UpdateOrder(ctx, &pb.Order{
		Id:          orderID,
		PaymentLink: paymentLink,
		Status:      "waiting",
	})

	return err
}
