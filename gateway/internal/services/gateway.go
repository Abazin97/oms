package services

import (
	"context"

	pbo "github.com/Abazin97/common/gen/go/order"
	pbs "github.com/Abazin97/common/gen/go/stock"
)

type OrdersGateway interface {
	CreateOrder(ctx context.Context, cr *pbo.CreateOrderRequest) (*pbo.Order, error)
	GetOrder(ctx context.Context, orderID string) (*pbo.Order, error)
}

type StockGateway interface {
	GetStock(ctx context.Context, cr *pbs.GetAvailabilityRequest) (*pbs.GetAvailabilityResponse, error)
}
