package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

func NewStorage(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("err")
	}
	return db, nil
}
