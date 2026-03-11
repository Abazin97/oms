package events

import (
	"orders/internal/domain/models"
	"time"
)

type OrderCreatedEvent struct {
	OrderID    string        `json:"order_id"`
	LotID      string        `json:"lot_id"`
	CustomerID string        `json:"customer_id"`
	From       time.Time     `json:"from"`
	To         time.Time     `json:"to"`
	Items      []models.Item `json:"items"`
	Status     string        `json:"status"`
}
