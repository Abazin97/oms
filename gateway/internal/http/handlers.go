package http

import (
	"encoding/json"
	"fmt"
	"gateway/internal"
	"gateway/internal/services"
	log "log/slog"
	"net/http"

	pb "github.com/Abazin97/common/gen/go/order"
)

type Handler interface {
	RegisterRoutes(mux *http.ServeMux)
}

type handler struct {
	gateway services.Gateway
}

func NewHandler(gateway services.Gateway) Handler {
	return &handler{gateway: gateway}
}

func (h *handler) RegisterRoutes(mux *http.ServeMux) {

	mux.Handle("/", http.FileServer(http.Dir("public")))

	mux.HandleFunc("POST /api/{customerID}/orders", h.createOrder)

	mux.HandleFunc("GET /api/{customerID}/orders/{orderID}", h.getOrder)
}

func (h *handler) createOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	customerID := r.PathValue("customerID")

	var items []*pb.ItemsQuantity
	if err := ReadJSON(r, &items); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	o, err := h.gateway.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerId: customerID,
		Items:      items,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	res := internal.CreateOrderRequest{
		Order:         o,
		RedirectToURL: fmt.Sprintf("http://localhost:50051/%s/orders/", o.Id),
	}

	WriteJSON(w, http.StatusOK, res)
}

func (h *handler) getOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("orderID")

	o, err := h.gateway.GetOrder(ctx, id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, o)
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
