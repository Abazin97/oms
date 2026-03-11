package events

import "time"

type StockStatusChangedEvent struct {
	ReservationID string    `json:"reservationID"`
	Status        string    `json:"status"`
	ChangedAt     time.Time `json:"changedAt"`
}
