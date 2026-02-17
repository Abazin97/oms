package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"payments/internal/gateway"
	"payments/internal/handlers"
	"payments/internal/services"
	"payments/internal/yookassa"
	"syscall"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	//cfg := config.MustLoad()
	log := setupLogger(envLocal)

	err := godotenv.Load(".env")
	if err != nil {
		log.Info("Error loading .env file")
	}

	grpcSrv := grpc.NewServer()

	ordersGateway, err := gateway.NewGateway("localhost:50052")
	if err != nil {
		log.Error("failed to connect grpc stock service", err)
		return
	}

	yooClient := yookassa.NewClient()

	paymentService := services.NewPaymentService(ordersGateway, yooClient)
	handlers.NewPaymentHandler(paymentService)
	handlers.NewGRPCHandler(grpcSrv, paymentService)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", 50053))
	if err != nil {
		log.Error("failed to listen: %v", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("gRPC server starting", slog.String("addr", l.Addr().String()))
		if err := grpcSrv.Serve(l); err != nil {
			log.Error("failed to serve: %v", err)
		}
	}()

	<-ctx.Done()
	ordersGateway.Close()
	grpcSrv.GracefulStop()
	log.Info("gRPC server stopped", slog.String("addr", l.Addr().String()))

	//if err := repo.Close(); err != nil {
	//	log.Error("failed to close repository: ", err)
	//}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
