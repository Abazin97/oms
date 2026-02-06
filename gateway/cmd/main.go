package main

import (
	delivery "gateway/internal/http"
	"gateway/internal/services/orders"
	"gateway/internal/services/stock"
	"log/slog"
	"net/http"
	"os"
)

var (
	serviceName = "services"
	httpAddr    = "localhost:4000"
	envLocal    = "local"
	envProd     = "prod"
)

func main() {
	const op = "gateway"
	log := setupLogger(envLocal)

	mux := http.NewServeMux()

	ordersGateway, err := orders.NewOrdersGateway("localhost:50051")
	if err != nil {
		log.Error("failed to connect grpc orders service", err)
		return
	}
	defer ordersGateway.Close()

	stocksGateway, err := stock.NewStockGateway("localhost:50052")
	if err != nil {
		log.Error("failed to connect grpc stock service", err)
		return
	}
	defer stocksGateway.Close()

	handler := delivery.NewHandler(ordersGateway, stocksGateway)
	handler.RegisterRoutes(mux)

	log.Info("starting HTTP server at %s", httpAddr)

	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Error("error starting HTTP server", err)
	}

	//log.Info(fmt.Sprintf("HTTP server started at: %s", httpAddr))
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
