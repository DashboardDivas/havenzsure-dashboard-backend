package server

import (
	"log"
	"net/http"
	"time"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/platform/middleware"
	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/shop"
	users "github.com/DashboardDivas/havenzsure-dashboard-backend/internal/user"
	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/workorder"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *Server) RegisterRoutes(db *pgxpool.Pool) http.Handler {
	router := chi.NewRouter()

	// Global middlewares
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.Timeout(5 * time.Second))

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

	// Protected routes (require authentication)
	// --- Shop route group ---
	shopRepo := shop.NewShopRepository(db)
	shopSvc := shop.NewService(shopRepo)
	shopHandler := shop.NewHandler(shopSvc)

	// -- User route group ---
	userRepo := users.NewUserRepository(db)
	var userSvc users.UserService
	if smtpSender, err := users.NewSMTPSenderFromEnv(); err != nil {
		log.Printf("WARNING: failed to initialize SMTP sender: %v; falling back to log-only sender", err)
		userSvc = users.NewService(userRepo)
	} else {
		userSvc = users.NewServiceWithEmailSender(userRepo, smtpSender)
	}
	userHandler := users.NewHandler(userSvc)

	// --- Me route group ---
	meSvc := users.NewMeService(userSvc)
	meHandler := users.NewMeHandler(meSvc)

	// --- WorkOrder route group ---
	workorderRepo := workorder.NewRepository(db)
	workorderSvc := workorder.NewService(workorderRepo)
	workorderHandler := workorder.NewHandler(workorderSvc)

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(userRepo)

	router.Group(func(r chi.Router) {
		// Apply authentication middleware
		// All routes inside this group require valid Firebase/GCIP ID Token
		r.Use(authMiddleware.Verify)

		// --- Me Routes (all authenticated users) ---
		// Note: /me routes always refer to the currently authenticated user
		// No user ID is needed in the URL
		r.Route("/me", func(sub chi.Router) {
			meHandler.RegisterRoutes(sub)
		})

		// --- Shop Routes (SuperAdmin + Admin only) ---
		r.Route("/shops", func(sub chi.Router) {
			sub.Use(middleware.RequireAdminOrAbove())
			shopHandler.RegisterRoutes(sub)
		})

		// --- User Routes (SuperAdmin + Admin only) ---
		// Note: Fine-grained permission checks are in service layer
		r.Route("/users", func(sub chi.Router) {
			sub.Use(middleware.RequireAdminOrAbove())
			userHandler.RegisterRoutes(sub)
		})

		// --- Work Order Routes (all authenticated users can access) ---
		// But with fine-grained permission control inside
		r.Route("/workorders", func(sub chi.Router) {
			sub.Use(middleware.EnforceShopScope())
			workorderHandler.RegisterRoutes(sub)
		})
	})

	return router
}
