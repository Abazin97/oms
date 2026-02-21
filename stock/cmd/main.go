package main

import (
	"context"
	"fmt"
	"gateway/discovery"
	"gateway/discovery/consul"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"stock/internal/handlers"
	"stock/internal/repository"
	"stock/internal/services"
	"stock/internal/tx"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

const (
	serviceName = "stock"
	grpcAddr    = "localhost:50052"
	consulAddr  = "localhost:8500"
	envLocal    = "local"
	envProd     = "prod"
)

func main() {

	log := setupLogger(envLocal)
	err := godotenv.Load(".env")
	if err != nil {
		log.Error("Error loading .env file")
	}

	url := os.Getenv("POSTGRES_URL")
	if url == "" {
		log.Error("url not set")
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

	db, err := repository.NewPostgresDB(url)
	if err != nil {
		log.Error("Error connecting to database")
		os.Exit(1)
	}
	defer db.Close()

	parkingRepo := repository.NewParkingRepository(db)
	reservationRepo := repository.NewSpotReservationRepository(db)

	txManager := tx.NewTxManager(db)
	stockService := services.NewStockService(txManager, parkingRepo, reservationRepo)
	handlers.NewGRPCHandler(grpcSrv, stockService)

	l, err := net.Listen("tcp", grpcAddr)
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
	grpcSrv.GracefulStop()
	log.Info("gRPC server stopped", slog.String("addr", l.Addr().String()))
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
