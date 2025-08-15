package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikeudacha/paybuy/cmd/api"
	"github.com/mikeudacha/paybuy/config"
	"github.com/mikeudacha/paybuy/db"
	"log"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
	pool, err := db.NewStorage(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	initStorage(pool)
	server := api.NewAPIServer(cfg.Host, pool)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

func initStorage(pool *pgxpool.Pool) {
	ctx, cansel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cansel()
	err := pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Couldn't connect to database")
	}
	log.Println("DB: Successfully connected!")
}
