package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
)

const (
	checkoutURL = "https://abazincloud.ddns.net"
	yookassaURL = "https://api.yookassa.ru/v3/payments"
	shopID      = "1249166"
)

type APIRequest struct {
	Request interface{} `json:"request"`
}

type APIResponse struct {
	Response json.RawMessage `json:"response"`
}

type CheckoutRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	Value          string `json:"value"`
	Currency       string `json:"currency"`
	Capture        bool   `json:"capture"`
	Description    string `json:"description"`
	ReturnURL      string `json:"return_url"`
}

type YouKassaRequest struct {
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Confirmation struct {
		Type      string `json:"type"`
		ReturnURL string `json:"return_url"`
	} `json:"confirmation"`
	PaymentMethodData struct {
		Type string `json:"type"`
	} `json:"payment_method_data"`
	Capture     bool   `json:"capture"`
	Description string `json:"description"`
}

func CreateRequest(cr *CheckoutRequest) YouKassaRequest {
	req := YouKassaRequest{}

	req.Amount.Value = cr.Value
	req.Amount.Currency = cr.Currency

	req.PaymentMethodData.Type = "bank_card"
	req.Confirmation.Type = "redirect"
	req.Confirmation.ReturnURL = cr.ReturnURL

	req.Capture = true
	req.Description = cr.Description

	return req
}

func main() {
	idempotencyKey := uuid.New().String()
	checkoutRequest := &CheckoutRequest{
		IdempotencyKey: idempotencyKey,
		Value:          "2.00",
		Currency:       "RUB",
		Capture:        true,
		Description:    "parking payment",
		ReturnURL:      checkoutURL,
	}

	req := APIRequest{Request: checkoutRequest}

	yooReq := CreateRequest(req.Request.(*CheckoutRequest))
	reqBody, _ := json.Marshal(yooReq)

	httpReq, err := http.NewRequest("POST", yookassaURL, bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}

	merchantPass := os.Getenv("MERCHANT_PASS")
	httpReq.SetBasicAuth(shopID, merchantPass)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", checkoutRequest.IdempotencyKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	apiResp := APIResponse{
		Response: json.RawMessage(body),
	}
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(apiResp.Response))
}
