package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"log/slog"

	"github.com/amillerrr/jwt-rbac-cors-app/internal/config"
	"github.com/amillerrr/jwt-rbac-cors-app/internal/handlers"
	"github.com/amillerrr/jwt-rbac-cors-app/internal/monitoring"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	config  *config.Config
	db      *sql.DB
	router  *http.ServeMux
	monitor *monitoring.Monitor
}

func New(cfg *config.Config, db *sql.DB) *Server {
	return NewWithMonitoring(cfg, db, nil)
}

func NewWithMonitoring(cfg *config.Config, db *sql.DB, monitor *monitoring.Monitor) *Server {
	s := &Server{
		config:  cfg,
		db:      db,
		router:  http.NewServeMux(),
		monitor: monitor,
	}

	s.setupRoutes()

	return s
}

func (s *Server) Start() error {
	if s.monitor != nil {
		s.monitor.Logger.Info("Server starting",
			"port", s.config.Server.Port,
			"cors_enabled", true,
		)
	} else {
		fmt.Printf("Server starting on port %s\n", s.config.Server.Port)
	}
	
	fmt.Println("CORS enabled - frontend can communicate with this backend")
	fmt.Println("Metrics endpoint: http://localhost:" + s.config.Server.Port + "/metrics")
	
	return http.ListenAndServe(":"+s.config.Server.Port, s.router)
}

func (s *Server) setupRoutes() {
	authHandler := handlers.NewAuthHandler(s.db, s.config.JWT.Secret, s.monitor.Logger)
	productHandler := handlers.NewProductHandler(s.db, s.monitor.Logger)
	adminHandler := handlers.NewAdminHandler(s.db, s.monitor.Logger)

	s.router.HandleFunc("/", corsMiddleware(s.serveStaticFiles))
	s.router.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("frontend/css/"))))
	s.router.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("frontend/js/"))))

	s.router.Handle("/metrics", promhttp.Handler())

	s.router.HandleFunc("/health", corsMiddleware(s.instrumentHandler("/health", s.healthHandler)))

	s.router.HandleFunc("/login", corsMiddleware(s.instrumentHandler("/login", authHandler.Login)))
	s.router.HandleFunc("/register", corsMiddleware(s.instrumentHandler("/register", authHandler.Register)))

	s.router.HandleFunc("/refresh", corsMiddleware(s.instrumentHandler("/refresh", authHandler.RefreshToken)))
	s.router.HandleFunc("/profile", corsMiddleware(s.instrumentHandler("/profile", authHandler.RequireAuth(authHandler.GetProfile))))

	s.router.HandleFunc("/products", corsMiddleware(s.instrumentHandler("/products", authHandler.RequireAuth(productHandler.GetProducts))))
	s.router.HandleFunc("/my-products", corsMiddleware(s.instrumentHandler("/my-products", authHandler.RequireAuth(productHandler.GetMyProducts))))

	s.router.HandleFunc("/admin", corsMiddleware(s.instrumentHandler("/admin", authHandler.RequireRole("admin", adminHandler.GetAdminData))))
	s.router.HandleFunc("/admin/stats", corsMiddleware(s.instrumentHandler("/admin/stats", authHandler.RequireRole("admin", adminHandler.GetSystemStats))))
	s.router.HandleFunc("/admin/users", corsMiddleware(s.instrumentHandler("/admin/users", authHandler.RequireRole("admin", adminHandler.GetAllUsers))))
}

func (s *Server) instrumentHandler(endpoint string, handler http.HandlerFunc) http.HandlerFunc {
	if s.monitor == nil {
		return handler
	}

	return s.monitor.HTTPMiddleware(handler)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status": "healthy", "message": "Backend server is running"}`)); err != nil {
		if s.monitor != nil && s.monitor.Logger != nil {
			s.monitor.Logger.Error("Failed to write health check response",
				slog.String("error", err.Error()),
			)
		}
		return
	}
}

func (s *Server) serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path == "/" {
		http.ServeFile(w, r, "frontend/index.html")
		return
	}

	filePath := "web" + r.URL.Path
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, serve the main HTML file (for SPA routing)
		http.ServeFile(w, r, "frontend/index.html")
		return
	}

	http.ServeFile(w, r, filePath)
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
