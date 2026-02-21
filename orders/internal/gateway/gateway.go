package gateway

import (
	"context"
	"orders/internal/domain/models"
	"time"

	"github.com/Abazin97/common/gen/go/stock"
)

type StockGateway interface {
	Reserve(ctx context.Context, lotID string, orderID string, from time.Time, to time.Time) (*stock.ReserveResponse, error)
}

type PaymentGateway interface {
	CreatePayment(ctx context.Context, orderID string, amount string, currency string) (*models.YouKassaResponse, error)
}
