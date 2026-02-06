package http

type getStockRequest struct {
	LotID string `json:"lot_id"`
	From  string `json:"from"`
	To    string `json:"to"`
}
