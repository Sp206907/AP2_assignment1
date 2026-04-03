package repository

import (
	"database/sql"
	"payment/internal/domain"
)

type postgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) PaymentRepository {
	return &postgresPaymentRepository{db: db}
}

func (r *postgresPaymentRepository) Create(p *domain.Payment) error {
	query := `INSERT INTO payments (id, order_id, transaction_id, amount, status)
              VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status)
	return err
}

func (r *postgresPaymentRepository) GetByOrderID(orderID string) (*domain.Payment, error) {
	query := `SELECT id, order_id, transaction_id, amount, status FROM payments WHERE order_id = $1`
	row := r.db.QueryRow(query, orderID)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
