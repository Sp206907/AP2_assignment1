package usecase

import (
	"errors"
	"payment/internal/domain"
	"payment/internal/repository"

	"github.com/google/uuid"
)

const MaxAmount int64 = 100000

type PaymentUseCase struct {
	repo repository.PaymentRepository
}

func NewPaymentUseCase(repo repository.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

func (uc *PaymentUseCase) Authorize(orderID string, amount int64) (*domain.Payment, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	status := domain.StatusAuthorized
	if amount > MaxAmount {
		status = domain.StatusDeclined
	}

	payment := &domain.Payment{
		ID:            uuid.NewString(),
		OrderID:       orderID,
		TransactionID: uuid.NewString(),
		Amount:        amount,
		Status:        status,
	}

	if err := uc.repo.Create(payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetByOrderID(orderID string) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(orderID)
}
