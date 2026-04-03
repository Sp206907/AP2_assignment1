package repository

import (
	"database/sql"
	"order/internal/domain"
)

type postgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) OrderRepository {
	return &postgresOrderRepository{db: db}
}

func (r *postgresOrderRepository) Create(order *domain.Order) error {
	query := `INSERT INTO orders (id, customer_id, item_name, amount, status, created_at, idempotency_key)
              VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''))`
	_, err := r.db.Exec(query, order.ID, order.CustomerID, order.ItemName,
		order.Amount, order.Status, order.CreatedAt, order.IdempotencyKey)
	return err
}

func (r *postgresOrderRepository) GetByIdempotencyKey(key string) (*domain.Order, error) {
	query := `SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE idempotency_key = $1`
	row := r.db.QueryRow(query, key)
	var o domain.Order
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}
func (r *postgresOrderRepository) GetByID(id string) (*domain.Order, error) {
	query := `SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var o domain.Order
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *postgresOrderRepository) UpdateStatus(id string, status string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}
func (r *postgresOrderRepository) GetRecent(limit int) ([]*domain.Order, error) {
	query := `SELECT id, customer_id, item_name, amount, status, created_at FROM orders ORDER BY created_at DESC LIMIT $1`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, nil
}
