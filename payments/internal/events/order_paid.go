package events

type OrderPaidEvent struct {
	OrderID  string `json:"order_id"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}
