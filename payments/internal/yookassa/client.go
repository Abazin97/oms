package yookassa

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"payments/internal/domain/models"
)

const yookassaURL = "https://api.yookassa.ru/v3/payments"

type Client struct {
	http        *http.Client
	yookassaURL string
	shopID      string
}

func NewClient() *Client {
	return &Client{
		http:        &http.Client{},
		yookassaURL: yookassaURL,
		shopID:      os.Getenv("SHOP_ID"),
	}
}

func (c *Client) CreatePayment(ctx context.Context, idempotenceKey string, req models.YouKassaRequest) (*models.YouKassaResponse, error) {
	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.yookassaURL, bytes.NewBuffer(body))
	httpReq.SetBasicAuth(c.shopID, os.Getenv("MERCHANT_PASS"))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", idempotenceKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var yooResp models.YouKassaResponse
	err = json.NewDecoder(resp.Body).Decode(&yooResp)
	return &yooResp, err
}
