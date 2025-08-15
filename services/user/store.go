package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikeudacha/paybuy/models"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, first_name, last_name, email, password, created_at FROM users WHERE email = $1`
	rows, err := s.pool.Query(context.Background(), query, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("user not found")
	}

	user, err := scanRowsIntoUser(rows)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) GetUserByID(id int) (*models.User, error) {
	query := `SELECT email FROM users WHERE id = $1`
	rows, err := s.pool.Query(context.Background(), query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("user not found")
	}

	user, err := scanRowsIntoUser(rows)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) CreateUser(user models.User) error {
	query := `INSERT INTO users (first_name, last_name, email, password) VALUES($1, $2, $3, $4)`
	_, err := s.pool.Exec(context.Background(), query, user.FirstName, user.LastName, user.Email, user.Password)
	if err != nil {
		return err
	}
	return nil
}

func scanRowsIntoUser(rows pgx.Rows) (*models.User, error) {
	user := new(models.User)

	err := rows.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}
