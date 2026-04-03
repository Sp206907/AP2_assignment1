package repository

import "payment/internal/domain"

type PaymentRepository interface {
	Create(payment *domain.Payment) error
	GetByOrderID(orderID string) (*domain.Payment, error)
}
