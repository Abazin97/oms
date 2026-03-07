package models

type OrderPaidEvent struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	OrderID  string `json:"orderID"`
	Status   string `json:"status"`
}
