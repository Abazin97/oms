package internal

import pb "github.com/Abazin97/common/gen/go/order"

type CreateOrderRequest struct {
	Order         *pb.Order `json:"order"`
	RedirectToURL string    `json:"redirect_to_url"`
}
