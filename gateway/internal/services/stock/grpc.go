package stock

import (
	"context"
	discovery2 "gateway/discovery"
	"gateway/internal/services"
	"log"

	pbs "github.com/Abazin97/common/gen/go/stock"
)

type stockGateway struct {
	service discovery2.Service
}

func NewStockGateway(service discovery2.Service) services.StockGateway {
	return &stockGateway{
		service: service,
	}
}

func (g *stockGateway) GetStock(ctx context.Context, cr *pbs.GetAvailabilityRequest) (*pbs.GetAvailabilityResponse, error) {
	conn, err := discovery2.ServiceConnection(context.Background(), "stock", g.service)
	if err != nil {
		log.Fatal("failed to dial server: ", err)
	}
	defer conn.Close()

	c := pbs.NewStockServiceClient(conn)
	return c.GetAvailability(ctx, cr)
}
