package server

import (
	"net/http"
	"time"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/shop"
	users "github.com/DashboardDivas/havenzsure-dashboard-backend/internal/user"
	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/workorder"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *Server) RegisterRoutes(db *pgxpool.Pool) http.Handler {
	router := chi.NewRouter()

	// All middlewares must be defined before routes on a mux
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(3 * time.Second))

	// Enables CORS so browser clients on other origins can call this API.
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// --- Shop route group ---
	repo := shop.NewShopRepository(db)
	svc := shop.NewService(repo)
	h := shop.NewHandler(svc)

	router.Route("/shops", func(sub chi.Router) {
		h.RegisterRoutes(sub)
	})

	// -- User route group ---
	userRepo := users.NewUserRepository(db)
	userSvc := users.NewService(userRepo)
	userHandler := users.NewHandler(userSvc)

	router.Route("/users", func(sub chi.Router) {
		userHandler.RegisterRoutes(sub)
	})

	// --- WorkOrder route group ---
	workorderRepo := workorder.NewRepository(db)
	workorderSvc := workorder.NewService(workorderRepo)
	workorderHandler := workorder.NewHandler(workorderSvc)

	router.Route("/workorders", func(sub chi.Router) {
		workorderHandler.RegisterRoutes(sub)
	})

	return router
}
