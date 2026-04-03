package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"order/internal/usecase"
)

type httpPaymentClient struct {
	client  *http.Client
	baseURL string
}

func NewHTTPPaymentClient(client *http.Client, baseURL string) usecase.PaymentClient {
	return &httpPaymentClient{client: client, baseURL: baseURL}
}

type paymentRequestBody struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type paymentResponseBody struct {
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
}

func (c *httpPaymentClient) Authorize(req usecase.PaymentRequest) (*usecase.PaymentResponse, error) {
	body, _ := json.Marshal(paymentRequestBody{OrderID: req.OrderID, Amount: req.Amount})

	resp, err := c.client.Post(c.baseURL+"/payments", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.New("payment service unavailable")
	}
	defer resp.Body.Close()

	var result paymentResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &usecase.PaymentResponse{
		Status:        result.Status,
		TransactionID: result.TransactionID,
	}, nil
}
