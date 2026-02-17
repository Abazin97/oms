package gateway

import "context"

type OrdersGateway interface {
	UpdateOrder(ctx context.Context, orderID string, status string) error
	Close() error
}
