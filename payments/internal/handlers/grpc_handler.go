package handlers

import (
	"context"
	"payments/internal/services"

	pb "github.com/Abazin97/common/gen/go/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	pb.UnimplementedPaymentServiceServer
	paymentService services.PaymentService
}

func NewGRPCHandler(grpcSrv *grpc.Server, service services.PaymentService) {
	handler := &serverAPI{paymentService: service}

	pb.RegisterPaymentServiceServer(grpcSrv, handler)
}

func (h *serverAPI) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	resp, err := h.paymentService.CreatePayment(ctx, req.OrderId, req.Amount, req.Currency)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "yookassa error: %v", err)
	}

	return &pb.CreatePaymentResponse{
		PaymentId:  resp.ID,
		PaymentUrl: resp.Confirmation.ConfirmationURL,
		Status:     resp.Status,
	}, nil
}
