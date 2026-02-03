package services

import (
	"context"

	pb "github.com/Abazin97/common/gen/go/order"
)

type Gateway interface {
	CreateOrder(ctx context.Context, cr *pb.CreateOrderRequest) (*pb.Order, error)
	GetOrder(ctx context.Context, orderID string) (*pb.Order, error)
	Close() error
}
