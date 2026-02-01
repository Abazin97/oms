package handlers

import (
	"context"
	log "log/slog"
	"orders/internal/domain/models"
	"orders/internal/services"

	pb "github.com/Abazin97/common/gen/go/order"
	"google.golang.org/grpc"
)

type serverAPI struct {
	pb.UnimplementedOrderServiceServer

	service services.OrdersService
}

func NewGRPCHandler(grpcSrv *grpc.Server, service services.OrdersService) {
	handler := &serverAPI{
		service: service,
	}
	pb.RegisterOrderServiceServer(grpcSrv, handler)
}

func (h *serverAPI) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.Order, error) {
	items := make([]models.Item, len(req.Items))
	for i, item := range req.Items {
		items[i] = models.Item{
			Id:       item.Id,
			Quantity: item.Quantity,
		}
	}

	createdOrder, err := h.service.CreateOrder(ctx, req.CustomerId, items)
	if err != nil {
		return nil, err
	}
	log.Info("CreateOrder request",
		"customerId", req.CustomerId,
		"itemsCount", len(req.Items),
	)

	pbItems := make([]*pb.Item, len(createdOrder.Items))
	for i, item := range createdOrder.Items {
		pbItems[i] = &pb.Item{
			Id:       item.Id,
			Quantity: item.Quantity,
			Price:    item.Price,
			Name:     item.Name,
		}
	}

	order := &pb.Order{
		Id:         createdOrder.Id,
		Status:     createdOrder.Status,
		Items:      pbItems,
		CustomerId: createdOrder.CustomerId,
	}

	return order, nil
}

func (h *serverAPI) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.Order, error) {
	o, err := h.service.GetOrder(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (h *serverAPI) UpdateOrder(ctx context.Context, req *pb.Order) (*pb.Order, error) {
	return h.service.UpdateOrder(ctx)
}
