package usecase

import (
	"errors"
	"order/internal/domain"
	"order/internal/repository"
	"time"

	"github.com/google/uuid"
)

type OrderUseCase struct {
	repo          repository.OrderRepository
	paymentClient PaymentClient
}

func NewOrderUseCase(repo repository.OrderRepository, paymentClient PaymentClient) *OrderUseCase {
	return &OrderUseCase{repo: repo, paymentClient: paymentClient}
}

func (uc *OrderUseCase) CreateOrder(customerID, itemName string, amount int64, idempotencyKey string) (*domain.Order, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	// Check idempotency
	if idempotencyKey != "" {
		existing, err := uc.repo.GetByIdempotencyKey(idempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	order := &domain.Order{
		ID:             uuid.NewString(),
		CustomerID:     customerID,
		ItemName:       itemName,
		Amount:         amount,
		Status:         domain.StatusPending,
		CreatedAt:      time.Now(),
		IdempotencyKey: idempotencyKey,
	}

	if err := uc.repo.Create(order); err != nil {
		return nil, err
	}

	resp, err := uc.paymentClient.Authorize(PaymentRequest{
		OrderID: order.ID,
		Amount:  order.Amount,
	})
	if err != nil {
		_ = uc.repo.UpdateStatus(order.ID, domain.StatusFailed)
		return nil, errors.New("payment service unavailable")
	}

	if resp.Status == "Authorized" {
		_ = uc.repo.UpdateStatus(order.ID, domain.StatusPaid)
		order.Status = domain.StatusPaid
	} else {
		_ = uc.repo.UpdateStatus(order.ID, domain.StatusFailed)
		order.Status = domain.StatusFailed
	}

	return order, nil
}

func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	return uc.repo.GetByID(id)
}
func (uc *OrderUseCase) GetRecentOrders(limit int) ([]*domain.Order, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be greater than 0")
	}
	return uc.repo.GetRecent(limit)
}

func (uc *OrderUseCase) CancelOrder(id string) error {
	order, err := uc.repo.GetByID(id)
	if err != nil {
		return errors.New("order not found")
	}
	if order.Status == domain.StatusPaid {
		return errors.New("paid orders cannot be cancelled")
	}
	if order.Status != domain.StatusPending {
		return errors.New("only pending orders can be cancelled")
	}
	return uc.repo.UpdateStatus(id, domain.StatusCancelled)
}
