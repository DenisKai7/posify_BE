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

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"posify-backend/internal/handler"
	"posify-backend/internal/middleware"
	"posify-backend/internal/repository"
	"posify-backend/internal/service"
	"posify-backend/pkg/database"
)

func main() {
	_ = godotenv.Load() // .env optional in production

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	log.Println("✓ Connected to Supabase PostgreSQL")

	log.Println("Running migrations...")
	if err := database.RunMigrations(context.Background(), pool); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("✓ Migrations complete")

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "db unreachable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth routes
	authRepo := repository.NewAuthRepo(pool)
	authSvc := service.NewAuthService(authRepo)
	authHandler := handler.NewAuthHandler(authSvc)

	// Sync routes
	syncRepo := repository.NewSyncRepo(pool)
	syncSvc := service.NewSyncService(syncRepo, pool)
	syncHandler := handler.NewSyncHandler(syncSvc)

	// Product management
	productRepo := repository.NewProductRepo(pool)
	productSvc := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productSvc)

	// Reports
	reportRepo := repository.NewReportRepo(pool)
	reportSvc := service.NewReportService(reportRepo)
	reportHandler := handler.NewReportHandler(reportSvc)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Protected sync endpoints — all authenticated roles
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Post("/sync/push", syncHandler.Push)
			r.Get("/sync/pull", syncHandler.Pull)
		})

		// Protected product list — all authenticated roles (for cashier pull)
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Get("/products", productHandler.List)
		})

		// Product management — OWNER & MANAGER only
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Use(middleware.RequireRole("OWNER", "MANAGER"))
			r.Post("/products", productHandler.Create)
			r.Put("/products/{id}", productHandler.Update)
			r.Delete("/products/{id}", productHandler.Delete)
		})

		// Reports — OWNER only
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Use(middleware.RequireRole("OWNER"))
			r.Get("/reports/summary", reportHandler.Summary)
			r.Get("/reports/sales-chart", reportHandler.SalesChart)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	fmt.Println("\nShutting down...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}
