package gateway

import "context"

type OrdersGateway interface {
	UpdateOrder(ctx context.Context, orderID string, paymentLink string) error
}
