package main

import (
	"context"
	"fmt"
	"gateway/discovery"
	"gateway/discovery/consul"
	delivery "gateway/internal/http"
	"gateway/internal/services/orders"
	"gateway/internal/services/stock"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var (
	serviceName = "gateway"
	httpAddr    = "localhost:4000"
	consulAddr  = "localhost:8500"
	envLocal    = "local"
	envProd     = "prod"
)

func main() {
	const op = "gateway"
	log := setupLogger(envLocal)

	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanseID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanseID, serviceName, httpAddr); err != nil {
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

	ordersGateway := orders.NewOrdersGateway(registry)
	stocksGateway := stock.NewStockGateway(registry)

	mux := http.NewServeMux()
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
