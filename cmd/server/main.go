package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amillerrr/jwt-rbac-cors-app/internal/config"
	"github.com/amillerrr/jwt-rbac-cors-app/internal/database"
	"github.com/amillerrr/jwt-rbac-cors-app/internal/monitoring"
	"github.com/amillerrr/jwt-rbac-cors-app/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	monitor, err := monitoring.NewMonitor(monitoring.Config{
		ServiceName:    "auth-app",
		ServiceVersion: "1.0.0",
		Environment:    getEnv("ENVIRONMENT", "development"),
		LogLevel:       slog.LevelInfo,
		LogFormat:      "json", // JSON logs are easier for Promtail to parse
		OTLPEndpoint:   getEnv("OTEL_ENDPOINT", "localhost:4318"), // Jaeger endpoint
		EnableMetrics:  true,
		EnableTracing:  true,
		EnableLogging:  true,
	})
	if err != nil {
		log.Fatalf("Failed to initialize monitoring: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		monitor.Logger.Info("Received shutdown signal, starting graceful shutdown...")
		cancel()
	}()

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		monitor.Logger.Error("Failed to connect to database", 
			slog.String("error", err.Error()),
		)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	instrumentedDB := database.NewInstrumentedDB(db, monitor.Metrics)

	monitor.Logger.Info("Database connection established successfully",
		slog.String("host", cfg.Database.Host),
		slog.Int("port", cfg.Database.Port),
	)

	go updateBusinessMetrics(ctx, instrumentedDB, monitor)

	srv := server.NewWithMonitoring(cfg, instrumentedDB, monitor)
	
	monitor.Logger.Info("Starting HTTP server",
		slog.String("port", cfg.Server.Port),
		slog.String("metrics_endpoint", "/metrics"),
	)

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		monitor.Logger.Info("Context cancelled, shutting down...")
	case err := <-serverErr:
		monitor.Logger.Error("Server error", slog.String("error", err.Error()))
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := monitor.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during monitoring shutdown: %v", err)
	}

	monitor.Logger.Info("Application shutdown complete")
}

func updateBusinessMetrics(ctx context.Context, db database.DB, monitor *monitoring.Monitor) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	updateMetrics(ctx, db, monitor)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			updateMetrics(ctx, db, monitor)
		}
	}
}

func updateMetrics(ctx context.Context, db database.DB, monitor *monitoring.Monitor) {
	defer monitor.TraceSpan(ctx, "update_business_metrics")()

	var totalUsers int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers); err != nil {
		monitor.LogError(ctx, "Failed to count total users", err)
	} else {
		monitor.Metrics.UsersTotal.Set(float64(totalUsers))
	}

	var activeUsers int
	query := "SELECT COUNT(*) FROM users WHERE last_login > NOW() - INTERVAL '24 hours'"
	if err := db.QueryRowContext(ctx, query).Scan(&activeUsers); err != nil {
		monitor.LogError(ctx, "Failed to count active users", err)
	} else {
		monitor.Metrics.UsersActive.Set(float64(activeUsers))
	}

	var totalProducts int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products WHERE is_active = true").Scan(&totalProducts); err != nil {
		monitor.LogError(ctx, "Failed to count products", err)
	} else {
		monitor.Metrics.ProductsTotal.Set(float64(totalProducts))
	}

	stats := db.Stats()
	monitor.Metrics.DBConnectionsOpen.Set(float64(stats.OpenConnections))

	monitor.Logger.Info("Updated business metrics",
		slog.Int("total_users", totalUsers),
		slog.Int("active_users", activeUsers),
		slog.Int("total_products", totalProducts),
		slog.Int("db_connections", stats.OpenConnections),
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
