package order

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikeudacha/paybuy/models"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateOrder(order models.Order) (int, error) {
	var id int
	query := `INSERT INTO orders (user_id, total, status, address) VALUES($1, $2, $3, $4) RETURNING id`
	err := s.pool.QueryRow(context.Background(), query, order.UserID, order.Total, order.Status, order.Address).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) CreateOrderItem(orderItem models.OrderItem) error {
	query := `INSERT INTO order_items (order_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)`
	_, err := s.pool.Exec(context.Background(), query, orderItem.OrderID, orderItem.ProductID, orderItem.Quantity, orderItem.Price)
	return err
}
