package events

import "time"

type OrderCreatedEvent struct {
	OrderID string    `json:"order_id"`
	LotID   string    `json:"lot_id"`
	From    time.Time `json:"from"`
	To      time.Time `json:"to"`
}
