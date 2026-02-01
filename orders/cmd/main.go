package main

import (
	"fmt"
	log "log/slog"
	"net"
	"orders/internal/handlers"
	"orders/internal/repository"
	"orders/internal/services"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	//cfg := config.MustLoad()

	err := godotenv.Load(".env")
	if err != nil {
		log.Info("Error loading .env file")
	}

	url := os.Getenv("POSTGRES_URL")
	if url == "" {
		log.Error("url not set")
	}

	grpcSrv := grpc.NewServer()

	repo, err := repository.NewPostgresRepository(url)
	if err != nil {
		log.Error("failed to connect to postgres repository: %v", err)
	}

	ordersService := services.NewOrdersService(repo)
	handlers.NewGRPCHandler(grpcSrv, ordersService)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051))
	if err != nil {
		log.Error("failed to listen: %v", err)
	}

	log.Info("gRPC server starting", log.String("addr", l.Addr().String()))

	if err := grpcSrv.Serve(l); err != nil {
		log.Error("failed to serve: %v", err)
	}

}
