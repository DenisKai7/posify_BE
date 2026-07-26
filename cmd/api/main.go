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

	// Start background scheduler
	scheduler := service.NewScheduler(pool)
	scheduler.Start()
	defer scheduler.Stop()

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://172.20.10.2:3000"},
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

	// Rate limiters
	authLimiter := middleware.RateLimit(10, time.Minute)    // 10 auth attempts/min
	paymentLimiter := middleware.RateLimit(30, time.Minute) // 30 payment ops/min

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authLimiter)
			r.Post("/auth/register", authHandler.Register)
			r.Post("/auth/login", authHandler.Login)
		})

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

		// Payments — QRIS
		paymentRepo := repository.NewPaymentRepo(pool)
		paymentSvc := service.NewPaymentService(paymentRepo)
		paymentHandler := handler.NewPaymentHandler(paymentSvc)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Use(paymentLimiter)
			r.Post("/payments/qris/generate", paymentHandler.GenerateQRIS)
			r.Get("/payments/qris/status/{orderID}", paymentHandler.GetStatus)
		})
		r.Post("/payments/qris/webhook", paymentHandler.Webhook) // public webhook (Perbaikan: hapus prefix /api/v1 ganda)

		// Customers & Loyalty
		customerRepo := repository.NewCustomerRepo(pool)
		customerSvc := service.NewCustomerService(customerRepo)
		customerHandler := handler.NewCustomerHandler(customerSvc)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Get("/customers", customerHandler.Search)
			r.Get("/customers/{id}", customerHandler.Get)
			r.Post("/customers", customerHandler.Create)
			r.Post("/customers/{id}/redeem", customerHandler.RedeemPoints)
		})

		// Inventory — stock adjustments & alerts
		stockRepo := repository.NewStockRepo(pool)
		stockSvc := service.NewStockService(stockRepo)
		stockHandler := handler.NewStockHandler(stockSvc)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Use(middleware.RequireRole("OWNER", "MANAGER"))
			r.Post("/inventory/adjust", stockHandler.Adjust)
			r.Get("/inventory/low-stock-alerts", stockHandler.LowStockAlerts)
			r.Get("/inventory/history/{productId}", stockHandler.History)
		})

		// Discounts
		discountRepo := repository.NewDiscountRepo(pool)
		discountSvc := service.NewDiscountService(discountRepo)
		discountHandler := handler.NewDiscountHandler(discountSvc)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Post("/discounts/validate", discountHandler.Validate)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Use(middleware.RequireRole("OWNER", "MANAGER"))
			r.Post("/discounts", discountHandler.Create)
			r.Get("/discounts", discountHandler.List)
		})

		// Shifts — all authenticated roles
		shiftRepo := repository.NewShiftRepo(pool)
		shiftSvc := service.NewShiftService(shiftRepo)
		shiftHandler := handler.NewShiftHandler(shiftSvc)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Get("/shifts/current", shiftHandler.GetCurrent)
			r.Post("/shifts/start", shiftHandler.Start)
			r.Post("/shifts/close", shiftHandler.Close)
		})

		// Payment Gateway Settings — OWNER & MANAGER
		pgRepo := repository.NewPaymentGatewayRepo(pool)
		pgSvc := service.NewPaymentGatewayService(pgRepo)
		pgHandler := handler.NewPaymentGatewayHandler(pgSvc)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Use(middleware.RequireRole("OWNER", "MANAGER"))
			r.Get("/settings/payment", pgHandler.Get)
			r.Put("/settings/payment", pgHandler.Update)
			r.Post("/settings/payment/test", pgHandler.Test)
		})

		// Midtrans QRIS Integration
		midtransSvc := service.NewMidtransService(pool, pgRepo, paymentRepo)
		midtransHandler := handler.NewMidtransHandler(midtransSvc)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Post("/payments/qris/charge", midtransHandler.Charge)
		})
		r.Post("/payments/midtrans/webhook", midtransHandler.Webhook) // public webhook

		// Membership Tiers — All Authenticated Users
		tierRepo := repository.NewTierRepo(pool)
		tierSvc := service.NewTierService(tierRepo)
		tierHandler := handler.NewTierHandler(tierSvc)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Get("/tiers", tierHandler.List)
		})

		// Membership Tiers Management — OWNER & MANAGER Only
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Use(middleware.RequireRole("OWNER", "MANAGER"))
			r.Put("/tiers/{id}", tierHandler.Update)
		})

		// Reports — OWNER & MANAGER
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)
			r.Use(middleware.RequireRole("OWNER", "MANAGER"))
			r.Get("/reports/summary", reportHandler.Summary)
			r.Get("/reports/sales-chart", reportHandler.SalesChart)
			r.Get("/reports/sales-summary", reportHandler.SalesSummary)
			r.Get("/reports/top-products", reportHandler.TopProducts)
			r.Get("/reports/payment-methods", reportHandler.PaymentMethods)
			r.Get("/reports/export/excel", reportHandler.ExportExcel)
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