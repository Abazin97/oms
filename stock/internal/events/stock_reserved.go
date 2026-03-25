package events

type StockReservedEvent struct {
	ReservationID string `json:"reservation_id"`
	OrderID       string `json:"order_id"`
	Status        string `json:"status"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
}
