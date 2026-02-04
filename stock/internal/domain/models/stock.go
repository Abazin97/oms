package models

type Stock struct {
	Id          string `json:"id"`
	Status      string `json:"status"`
	CustomerId  string `json:"customer_id"`
	Items       []Item `json:"items"`
	PaymentLink string `json:"payment_link"`
}

type Item struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Quantity int32  `json:"quantity"`
	Price    string `json:"price"`
}
