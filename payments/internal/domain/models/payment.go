package models

type YouKassaRequest struct {
	Amount       Amount            `json:"amount"`
	Confirmation Confirmation      `json:"confirmation"`
	Capture      bool              `json:"capture"`
	Description  string            `json:"description"`
	Metadata     map[string]string `json:"metadata"`
}

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type Confirmation struct {
	Type      string `json:"type"`
	ReturnURL string `json:"return_url"`
}

type YouKassaResponse struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Amount       Amount `json:"amount"`
	Confirmation struct {
		ConfirmationURL string `json:"confirmation_url"`
	} `json:"confirmation"`
}

type YouKassaNotification struct {
	Type   string `json:"type"`
	Event  string `json:"event"`
	Object struct {
		ID       string            `json:"id"`
		Status   string            `json:"status"`
		Paid     bool              `json:"paid"`
		Amount   Amount            `json:"amount"`
		Metadata map[string]string `json:"metadata"`
	} `json:"object"`
}
