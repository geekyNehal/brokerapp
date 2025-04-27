package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"brokerapp/internal/config"
	"brokerapp/internal/db"
	"brokerapp/internal/holdings"
	"brokerapp/internal/orderbook"
	"brokerapp/internal/positions"
	"brokerapp/internal/user"
	"brokerapp/pkg/authmiddleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create database connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// Initialize database with circuit breaker
	mysqlDB, err := db.NewMySQL(dsn)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer mysqlDB.Close()

	// Initialize repositories
	userRepo := user.NewMySQLRepository(mysqlDB)

	// Initialize services
	userService := user.NewService(userRepo, cfg.JWTSecret)

	// Initialize handlers
	userHandler := user.NewHandler(userService)
	holdingsHandler := holdings.NewHandler(mysqlDB)
	orderbookHandler := orderbook.NewHandler(mysqlDB)
	positionsHandler := positions.NewHandler(mysqlDB)

	// Initialize router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Public routes
	r.Route("/api", func(r chi.Router) {
		// User routes
		r.Post("/signup", userHandler.SignUp)
		r.Post("/login", userHandler.Login)
		r.Post("/refresh", userHandler.RefreshToken)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authmiddleware.AuthMiddleware(cfg.JWTSecret))
			r.Get("/profile", userHandler.GetProfile)
			holdingsHandler.RegisterRoutes(r)
			r.Get("/orderbook", orderbookHandler.GetOrderbook)
			r.Get("/positions", positionsHandler.GetPositions)
		})
	})

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Create shutdown context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
