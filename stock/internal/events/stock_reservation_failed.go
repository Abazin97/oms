package events

type StockReservationFailedEvent struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}
