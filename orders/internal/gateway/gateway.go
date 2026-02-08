package gateway

import (
	"context"
	"time"

	"github.com/Abazin97/common/gen/go/stock"
)

type StockGateway interface {
	Reserve(context.Context, string, string, time.Time, time.Time) (*stock.ReserveResponse, error)
	Close() error
}
