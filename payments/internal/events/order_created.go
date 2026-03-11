package events

type OrderCreatedEvent struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
	Status     string `json:"status"`
}
