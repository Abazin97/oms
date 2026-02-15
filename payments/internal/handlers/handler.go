package handlers

import "net/http"

type PaymentHandler struct {
}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{}
}

func (h *PaymentHandler) HandleYouKassaWebHook(w http.ResponseWriter, r *http.Request) {
	
}
