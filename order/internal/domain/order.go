package domain

import "time"

type Order struct {
	ID             string
	CustomerID     string
	ItemName       string
	Amount         int64
	Status         string
	CreatedAt      time.Time
	IdempotencyKey string
}

const (
	StatusPending   = "Pending"
	StatusPaid      = "Paid"
	StatusFailed    = "Failed"
	StatusCancelled = "Cancelled"
)
