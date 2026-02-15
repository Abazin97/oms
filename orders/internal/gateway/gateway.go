package gateway

import (
	"context"
	"time"

	"github.com/Abazin97/common/gen/go/stock"
)

type StockGateway interface {
	Reserve(ctx context.Context, lotID string, orderID string, from time.Time, to time.Time) (*stock.ReserveResponse, error)
	Close() error
}
