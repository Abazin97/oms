package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"payments/internal/domain/models"
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
	mux.HandleFunc("/api/payment/notifications", h.HandleYouKassaWebHook)
}

func (h *paymentHandler) HandleYouKassaWebHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	log.Println("yookassa webhook: ", string(body))

	var notification models.YouKassaNotification

	if err := json.Unmarshal(body, &notification); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if err := h.paymentService.HandleYouKassaWebHook(r.Context(), notification); err != nil {
		http.Error(w, "cannot handle webhook", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
