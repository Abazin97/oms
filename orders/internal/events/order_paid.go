package events

type OrderPaidEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}
