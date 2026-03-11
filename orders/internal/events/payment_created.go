package events

type PaymentCreatedEvent struct {
	OrderID    string `json:"order_id"`
	PaymentID  string `json:"payment_id"`
	PaymentURL string `json:"payment_link"`
	Status     string `json:"status"`
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
}
