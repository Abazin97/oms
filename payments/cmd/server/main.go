//package main
//
//import (
//	"encoding/json"
//	"fmt"
//	"io"
//	"log"
//	"net/http"
//)
//
//// APIResponse хранит данные запроса
//type APIResponse struct {
//	Response interface{} `json:"response"`
//}
//
//func main() {
//	http.HandleFunc("/api/payment", func(w http.ResponseWriter, r *http.Request) {
//		if r.Method != http.MethodPost {
//			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//			return
//		}
//
//		body, err := io.ReadAll(r.Body)
//		if err != nil {
//			http.Error(w, "cannot read body", http.StatusBadRequest)
//			return
//		}
//		defer r.Body.Close()
//
//		fmt.Println("Received body:")
//		fmt.Println(string(body))
//
//		var apiResp APIResponse
//		if len(body) > 0 {
//			if err := json.Unmarshal(body, &apiResp); err != nil {
//				fmt.Println("Failed to parse JSON:", err)
//			} else {
//				fmt.Println("Parsed response:", apiResp.Response)
//			}
//		} else {
//			fmt.Println("Empty body")
//		}
//		w.Header().Set("Content-Type", "application/json")
//		w.WriteHeader(http.StatusOK)
//		resp := map[string]string{"status": "ok"}
//		json.NewEncoder(w).Encode(resp)
//	})
//
//	log.Println("Server listening on :8080")
//	log.Fatal(http.ListenAndServe(":8080", nil))
//}

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

const (
	yookassaURL = "https://api.yookassa.ru/v3/payments"
	shopID      = "1249166"
)

type YouKassaRequest struct {
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`

	PaymentMethodData struct {
		Type string `json:"type"`
	} `json:"payment_method_data"`

	Confirmation struct {
		Type      string `json:"type"`
		ReturnURL string `json:"return_url"`
	} `json:"confirmation"`

	Capture     bool   `json:"capture"`
	Description string `json:"description"`
}

func main() {
	http.HandleFunc("/api/payment", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		log.Println("Received from Postman:")
		log.Println(string(body))

		var req YouKassaRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		idempotenceKey := uuid.New().String()

		yooReqBody, _ := json.Marshal(req)

		httpReq, err := http.NewRequest("POST", yookassaURL, bytes.NewBuffer(yooReqBody))
		if err != nil {
			http.Error(w, "cannot build request", http.StatusInternalServerError)
			return
		}

		merchantPass := os.Getenv("MERCHANT_PASS")
		httpReq.SetBasicAuth(shopID, merchantPass)
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Idempotence-Key", idempotenceKey)

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			http.Error(w, "yookassa request failed", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		yooRespBody, _ := io.ReadAll(resp.Body)

		log.Println("YooKassa response:")
		log.Println(string(yooRespBody))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(yooRespBody)
	})

	//http.HandleFunc("/api/payment/notifications", func(w http.ResponseWriter, r *http.Request) {
	//	if r.Method != http.MethodPost {
	//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	//		return
	//	}
	//	body, err := io.ReadAll(r.Body)
	//	if err != nil {
	//		http.Error(w, "cannot read body", http.StatusBadRequest)
	//		return
	//	}
	//
	//	defer r.Body.Close()
	//	log.Println(string(body))
	//
	//	var notification struct {
	//		Type   string `json:"type"`
	//		Event  string `json:"event"`
	//		Object struct {
	//			ID     string `json:"id"`
	//			Status string `json:"status"`
	//			Paid   bool   `json:"paid"`
	//		} `json:"object"`
	//	}
	//
	//	if err := json.Unmarshal(body, &notification); err != nil {
	//		http.Error(w, "bad json", http.StatusBadRequest)
	//		return
	//	}
	//
	//	w.WriteHeader(http.StatusOK)
	//	w.Write([]byte(`{"status":"ok"}`))
	//})
	//
	//log.Println("Server listening on :4001")
	//log.Fatal(http.ListenAndServe(":4001", nil))
}
