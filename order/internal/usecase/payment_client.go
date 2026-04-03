package usecase

type PaymentRequest struct {
	OrderID string
	Amount  int64
}

type PaymentResponse struct {
	Status        string
	TransactionID string
}

type PaymentClient interface {
	Authorize(req PaymentRequest) (*PaymentResponse, error)
}
