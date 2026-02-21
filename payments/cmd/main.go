package main

import (
	"context"
	"fmt"
	"gateway/discovery"
	"gateway/discovery/consul"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"payments/internal/gateway"
	"payments/internal/handlers"
	"payments/internal/services"
	"payments/internal/yookassa"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

const (
	serviceName = "payments"
	grpcAddr    = "localhost:50053"
	consulAddr  = "localhost:8500"
	envLocal    = "local"
	envProd     = "prod"
)

func main() {
	//cfg := config.MustLoad()
	log := setupLogger(envLocal)

	err := godotenv.Load(".env")
	if err != nil {
		log.Info("Error loading .env file")
	}

	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanseID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanseID, serviceName, grpcAddr); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanseID, serviceName); err != nil {
				log.Error(fmt.Sprintf("failed to health check"))
			}

			time.Sleep(5 * time.Second)
		}
	}()

	defer registry.Deregister(ctx, instanseID, serviceName)

	grpcSrv := grpc.NewServer()

	ordersGateway := gateway.NewGateway(registry)

	yooClient := yookassa.NewClient()

	paymentService := services.NewPaymentService(ordersGateway, yooClient)
	handlers.NewGRPCHandler(grpcSrv, paymentService)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", 50053))
	if err != nil {
		log.Error("failed to listen: %v", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	httpHandler := handlers.NewPaymentHandler(paymentService)
	httpHandler.RegisterRoutes(mux)

	httpSrv := &http.Server{
		Addr:    "localhost:4001",
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("gRPC server starting", slog.String("addr", l.Addr().String()))
		if err := grpcSrv.Serve(l); err != nil {
			log.Error("failed to serve: %v", err)
		}
	}()

	go func() {
		log.Info("http server starting :4001")
		if err := httpSrv.ListenAndServe(); err != nil {
			log.Error("failed to serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down...")
	grpcSrv.GracefulStop()
	httpSrv.Shutdown(context.Background())
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
