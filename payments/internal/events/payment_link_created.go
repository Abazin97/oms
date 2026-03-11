package events

type PaymentLinkCreatedEvent struct {
	OrderID     string `json:"order_id"`
	PaymentLink string `json:"payment_link"`
}
