package internal

import (
	"time"

	pb "github.com/Abazin97/common/gen/go/order"
)

type CreateOrderHTTPRequest struct {
	LotID string              `json:"lot_id"`
	From  time.Time           `json:"from"`
	To    time.Time           `json:"to"`
	Items []*pb.ItemsQuantity `json:"items"`
}

type CreateOrderResponse struct {
	Order         *pb.Order `json:"order"`
	RedirectToURL string    `json:"redirect_to_url"`
}
