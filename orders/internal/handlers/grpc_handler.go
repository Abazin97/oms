package handlers

import (
	"context"
	log "log/slog"
	"orders/internal/domain/models"
	"orders/internal/services"

	pb "github.com/Abazin97/common/gen/go/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

	//by Artem Topskiy
	//by Leonid Chizas

	createdOrder, err := h.service.CreateOrder(ctx, req.Id, req.CustomerId, req.From.AsTime(), req.To.AsTime(), items)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to create order")
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
		Id:          createdOrder.Id,
		Status:      createdOrder.Status,
		Items:       pbItems,
		CustomerId:  createdOrder.CustomerId,
		PaymentLink: createdOrder.PaymentLink,
	}

	return order, nil
}

func (h *serverAPI) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.Order, error) {
	order, err := h.service.GetOrder(ctx, req.Id)

	if err != nil {
		return nil, status.Error(codes.NotFound, "Order not found")
	}

	items := make([]*pb.Item, 0, len(order.Items))
	for _, it := range order.Items {
		items = append(items, &pb.Item{
			Id:       it.Id,
			Name:     it.Name,
			Quantity: it.Quantity,
			Price:    it.Price,
		})
	}
	// for multiple orders per customer

	//orders := make([]*pb.Order, 0, len(accountOrders))
	//for _, o := range accountOrders {
	//	items := make([]*pb.Item, 0, len(o.Items))
	//	for _, it := range o.Items {
	//		items = append(items, &pb.Item{
	//			Id:       it.Id,
	//			Quantity: it.Quantity,
	//			Price:    it.Price,
	//			Name:     it.Name,
	//		})
	//	}
	//	orders = append(orders, &pb.Order{
	//		Id:         o.Id,
	//		CustomerId: o.CustomerId,
	//		Status:     o.Status,
	//		Items:      []*pb.Item{},
	//	})
	//}

	return &pb.Order{
		Id:          order.Id,
		Status:      order.Status,
		CustomerId:  order.CustomerId,
		PaymentLink: order.PaymentLink,
		Items:       items,
	}, nil
}

func (h *serverAPI) UpdateOrder(ctx context.Context, o *pb.UpdateOrderStatusRequest) (*emptypb.Empty, error) {
	err := h.service.UpdateOrder(ctx, o.OrderId, o.Status)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Failed to update, order not found")
	}

	return &emptypb.Empty{}, nil
}
