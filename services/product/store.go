package product

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikeudacha/paybuy/models"
	"strings"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) GetProducts() ([]*models.Product, error) {
	query := `SELECT * FROM products`
	rows, err := s.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	products := make([]*models.Product, 0)
	for rows.Next() {
		p, err := scanRowsIntoProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (s *Store) CreateProduct(product models.CreateProductPayload) error {
	query := `INSERT INTO products (name, price, image, description, quantity) VALUES($1, $2, $3, $4, $5)`
	_, err := s.pool.Exec(context.Background(), query, product.Name, product.Price, product.Image, product.Description, product.Quantity)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetProductByID(productID int) (*models.Product, error) {
	query := `SELECT * FROM products WHERE id = $1`
	rows, err := s.pool.Query(context.Background(), query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanRowsIntoProduct(rows)
	}

	return nil, nil
}
func (s *Store) UpdateProduct(product models.Product) error {
	query := `UPDATE products SET name = $1,price = $2,image = $3,description = $4,quantity = $5 WHERE id = $6`
	_, err := s.pool.Exec(context.Background(), query,
		product.Name,
		product.Price,
		product.Image,
		product.Description,
		product.Quantity,
		product.ID,
	)

	return err
}

func (s *Store) GetProductsByID(productIDs []int) ([]models.Product, error) {
	if len(productIDs) == 0 {
		return []models.Product{}, nil
	}
	placeholders := strings.Repeat(",$", len(productIDs))
	query := fmt.Sprintf("SELECT * FROM products WHERE id IN ($1,%s)", placeholders)

	args := make([]interface{}, len(productIDs))
	for i, v := range productIDs {
		args[i] = v
	}

	rows, err := s.pool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		p, err := scanRowsIntoProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, *p)
	}

	return products, nil
}

func scanRowsIntoProduct(rows pgx.Rows) (*models.Product, error) {
	product := new(models.Product)

	err := rows.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Image,
		&product.Price,
		&product.Quantity,
		&product.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return product, nil
}
