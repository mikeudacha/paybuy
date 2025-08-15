package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikeudacha/paybuy/services/auth"
	"github.com/mikeudacha/paybuy/services/cart"
	"github.com/mikeudacha/paybuy/services/order"
	"github.com/mikeudacha/paybuy/services/product"
	"github.com/mikeudacha/paybuy/services/user"
)

type APIServer struct {
	addr string
	db   *pgxpool.Pool
}

func NewAPIServer(addr string, db *pgxpool.Pool) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()

	blacklistStore := auth.NewBlacklistStore(s.db)

	blacklistStore.CleanupExpiredTokensPeriodically(1 * time.Hour)

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore, blacklistStore)
	userHandler.RegisterRoutes(router)

	productStore := product.NewStore(s.db)
	productHandler := product.NewHandler(productStore, userStore, blacklistStore)

	orderStore := order.NewStore(s.db)

	cartHandler := cart.NewHandler(productStore, orderStore, userStore, blacklistStore)
	cartHandler.RegisterRoutes(router)

	productHandler.RegisterRoutes(router)

	log.Println("Server run in: ", s.addr)

	return http.ListenAndServe(s.addr, router)
}
