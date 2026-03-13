package events

type StockReservationFailedEvent struct {
	OrderID string `json:"order_id"`
	LotID   string `json:"lot_id"`
	Reason  string `json:"reason"`
}
