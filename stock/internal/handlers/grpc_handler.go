package handlers

import (
	"context"
	"stock/internal/services"

	pb "github.com/Abazin97/common/gen/go/stock"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	pb.UnimplementedStockServiceServer
	service services.StockService
}

func NewGRPCHandler(grpcSrv *grpc.Server, service services.StockService) {
	handler := &serverAPI{
		service: service,
	}
	pb.RegisterStockServiceServer(grpcSrv, handler)
}

func (s *serverAPI) GetAvailability(ctx context.Context, req *pb.GetAvailabilityRequest) (*pb.GetAvailabilityResponse, error) {
	id, err := uuid.Parse(req.LotId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid lot_id")
	}

	if req.From == nil || req.To == nil {
		return nil, status.Error(codes.InvalidArgument, "from and to are required")
	}

	from := req.From.AsTime()
	to := req.To.AsTime()

	if !to.After(from) {
		return nil, status.Error(codes.InvalidArgument, "to must be after from")
	}
	inStock, err := s.service.GetAvailability(ctx, id, from, to)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetAvailabilityResponse{
		Available: inStock,
	}, nil
}
