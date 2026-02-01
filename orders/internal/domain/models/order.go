package models

type Order struct {
	Id          string
	Status      string
	CustomerId  string
	Items       []Item
	PaymentLink string
}

type Item struct {
	Id       string
	Name     string
	Quantity int32
	Price    string
}
