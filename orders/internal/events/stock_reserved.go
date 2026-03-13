package events

import "time"

type StockReservedEvent struct {
	ReservationID string    `json:"reservation_id"`
	OrderID       string    `json:"order_id"`
	ParkingSpotID string    `json:"parking_spot_id"`
	Status        string    `json:"status"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
}
