package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"payments/internal/domain/models"
	"payments/internal/services"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentHandler interface {
	RegisterRoutes(mux *http.ServeMux)
}

type paymentHandler struct {
	paymentService services.PaymentService
	channel        *amqp.Channel
}

func NewPaymentHandler(service services.PaymentService, channel *amqp.Channel) PaymentHandler {
	return &paymentHandler{
		paymentService: service,
		channel:        channel}
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

	//if notification.Object.Status == "succeeded" {
	//	event := models.OrderPaidEvent{
	//		OrderID:  notification.Object.Metadata["orderId"],
	//		Amount:   notification.Object.Amount.Value,
	//		Currency: notification.Object.Amount.Currency,
	//		Status:   "paid",
	//	}
	//
	//	resp, err := json.Marshal(event)
	//	if err != nil {
	//		http.Error(w, "internal error", http.StatusInternalServerError)
	//		return
	//	}
	//
	//	err = h.channel.PublishWithContext(
	//		r.Context(),
	//		rabbitmq.OrderExchange,
	//		rabbitmq.OrderPaidEvent,
	//		false,
	//		false,
	//		amqp.Publishing{
	//			ContentType:  "application/json",
	//			Body:         resp,
	//			DeliveryMode: amqp.Persistent,
	//		})
	//	if err != nil {
	//		http.Error(w, "internal error", http.StatusInternalServerError)
	//		return
	//	}
	//
	//	log.Println("message published: order.paid", event.OrderID)
	//}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
