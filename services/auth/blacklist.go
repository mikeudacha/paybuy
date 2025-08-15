package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BlacklistStore struct {
	db *pgxpool.Pool
}

func NewBlacklistStore(db *pgxpool.Pool) *BlacklistStore {
	return &BlacklistStore{db: db}
}

func (s *BlacklistStore) AddToBlacklist(token string, userID int, expiresAt time.Time) error {
	query := `
		INSERT INTO blacklisted_tokens (token, user_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (token) DO NOTHING
	`

	_, err := s.db.Exec(context.Background(), query, token, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

func (s *BlacklistStore) IsBlacklisted(token string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM blacklisted_tokens 
			WHERE token = $1 AND expires_at > NOW()
		)
	`

	var exists bool
	err := s.db.QueryRow(context.Background(), query, token).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if token is blacklisted: %w", err)
	}

	return exists, nil
}

func (s *BlacklistStore) CleanupExpiredTokens() error {
	query := `DELETE FROM blacklisted_tokens WHERE expires_at <= NOW()`

	_, err := s.db.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	return nil
}

func (s *BlacklistStore) CleanupExpiredTokensPeriodically(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := s.CleanupExpiredTokens(); err != nil {
				fmt.Printf("Failed to cleanup expired tokens: %v\n", err)
			}
		}
	}()
}
