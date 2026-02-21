package http

import (
	"encoding/json"
	"gateway/internal"
	"gateway/internal/services"
	log "log/slog"
	"net/http"
	"time"

	pbo "github.com/Abazin97/common/gen/go/order"
	pbs "github.com/Abazin97/common/gen/go/stock"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler interface {
	RegisterRoutes(mux *http.ServeMux)
}

type handler struct {
	ordersGateway services.OrdersGateway
	stockGateway  services.StockGateway
}

func NewHandler(ordersGateway services.OrdersGateway, stockGateway services.StockGateway) Handler {
	return &handler{ordersGateway: ordersGateway, stockGateway: stockGateway}
}

func (h *handler) RegisterRoutes(mux *http.ServeMux) {

	//mux.Handle("/", http.FileServer(http.Dir("public")))

	mux.HandleFunc("POST /api/{customerID}/orders", h.createOrder)

	mux.HandleFunc("GET /api/{customerID}/orders/{orderID}", h.getOrder)

	mux.HandleFunc("GET /api/stock", h.getStock)
}

func (h *handler) createOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	customerID := r.PathValue("customerID")

	var req internal.CreateOrderHTTPRequest
	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	o, err := h.ordersGateway.CreateOrder(ctx, &pbo.CreateOrderRequest{
		CustomerId: customerID,
		Items:      req.Items,
		Id:         req.LotID,
		From:       timestamppb.New(req.From),
		To:         timestamppb.New(req.To),
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	res := internal.CreateOrderResponse{
		Order: o,
	}

	WriteJSON(w, http.StatusOK, res)
}

func (h *handler) getOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("orderID")

	o, err := h.ordersGateway.GetOrder(ctx, id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, o)
}

func (h *handler) getStock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req getStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	id := req.LotID
	if _, err := uuid.Parse(id); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid lotID")
		return
	}

	fromTime, err := time.Parse(time.RFC3339, req.From)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
	}

	toTime, err := time.Parse(time.RFC3339, req.To)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
	}

	fromProto := timestamppb.New(fromTime)
	toProto := timestamppb.New(toTime)
	s, err := h.stockGateway.GetStock(ctx, &pbs.GetAvailabilityRequest{
		LotId: id,
		From:  fromProto,
		To:    toProto,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, pbs.GetAvailabilityResponse{
		Available: s.Available,
	})
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Info("error writing JSON: %s", err)
	}
}

func ReadJSON(r *http.Request, data any) error {
	return json.NewDecoder(r.Body).Decode(data)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}
