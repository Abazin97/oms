package events

import "time"

type StockReleasedEvent struct {
	ReservationID string    `json:"reservation_id"`
	OrderID       string    `json:"order_id"`
	Reason        string    `json:"reason"`
	ReleasedAt    time.Time `json:"released_at"`
}
