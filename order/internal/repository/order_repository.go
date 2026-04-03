package repository

import "order/internal/domain"

type OrderRepository interface {
	Create(order *domain.Order) error
	GetByID(id string) (*domain.Order, error)
	UpdateStatus(id string, status string) error
	GetByIdempotencyKey(key string) (*domain.Order, error)
	GetRecent(limit int) ([]*domain.Order, error)
}
