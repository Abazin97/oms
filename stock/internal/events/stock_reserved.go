package events

import "time"

type StockReservedEvent struct {
	ReservationID string    `json:"reservationID"`
	OrderID       string    `json:"orderID"`
	Status        string    `json:"status"`
	StartsAt      time.Time `json:"startsAt"`
	EndsAt        time.Time `json:"endsAt"`
	CreatedAt     time.Time `json:"createdAt"`
}
