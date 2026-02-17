package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"payments/internal/services"
)

type PaymentHandler interface {
	RegisterRoutes(mux *http.ServeMux)
}

type paymentHandler struct {
	paymentService services.PaymentService
}

func NewPaymentHandler(service services.PaymentService) PaymentHandler {
	return &paymentHandler{paymentService: service}
}

func (h *paymentHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/payment/notifications", h.HandleYouKassaWebHook)
}

func (h *paymentHandler) HandleYouKassaWebHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	log.Println(string(body))

	var notification struct {
		Type   string `json:"type"`
		Event  string `json:"event"`
		Object struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Paid   bool   `json:"paid"`
		} `json:"object"`
	}

	if err := json.Unmarshal(body, &notification); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
