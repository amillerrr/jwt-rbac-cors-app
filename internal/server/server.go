package server

import (
	"database/sql"
	"fmt"
	"os"
	"net/http"

	"github.com/amillerrr/jwt-rbac-cors-app/internal/config"
	"github.com/amillerrr/jwt-rbac-cors-app/internal/handlers"
)

// Server represents the HTTP server with its dependencies
type Server struct {
	config *config.Config
	db     *sql.DB
	router *http.ServeMux
}

// New creates a new server instance with all dependencies
func New(cfg *config.Config, db *sql.DB) *Server {
	s := &Server{
		config: cfg,
		db:     db,
		router: http.NewServeMux(),
	}

	// Initialize routes
	s.setupRoutes()

	return s
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	fmt.Printf("Server starting on port %s\n", s.config.Server.Port)
	fmt.Println("CORS enabled - frontend can communicate with this backend")
	
	return http.ListenAndServe(":"+s.config.Server.Port, s.router)
}

// setupRoutes configures all HTTP routes with appropriate middleware
func (s *Server) setupRoutes() {
	// Initialize handlers with dependencies
	authHandler := handlers.NewAuthHandler(s.db, s.config.JWT.Secret)
	productHandler := handlers.NewProductHandler(s.db)
	adminHandler := handlers.NewAdminHandler(s.db)

	// Serve static files (frontend)
	s.router.HandleFunc("/", corsMiddleware(s.serveStaticFiles))
	// s.router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static/"))))
	s.router.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("frontend/css/"))))
	s.router.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("frontend/js/"))))

	// Public routes (no authentication required)
	s.router.HandleFunc("/health", corsMiddleware(s.healthHandler))
	s.router.HandleFunc("/login", corsMiddleware(authHandler.Login))
	s.router.HandleFunc("/register", corsMiddleware(authHandler.Register))

	// Authentication routes
	s.router.HandleFunc("/refresh", corsMiddleware(authHandler.RefreshToken))
	s.router.HandleFunc("/profile", corsMiddleware(authHandler.RequireAuth(authHandler.GetProfile)))

	// Product routes (authentication required)
	s.router.HandleFunc("/products", corsMiddleware(authHandler.RequireAuth(productHandler.GetProducts)))
	s.router.HandleFunc("/my-products", corsMiddleware(authHandler.RequireAuth(productHandler.GetMyProducts)))

	// Admin routes (authentication + admin role required)
	s.router.HandleFunc("/admin", corsMiddleware(authHandler.RequireRole("admin", adminHandler.GetAdminData)))
	s.router.HandleFunc("/admin/stats", corsMiddleware(authHandler.RequireRole("admin", adminHandler.GetSystemStats)))
	s.router.HandleFunc("/admin/users", corsMiddleware(authHandler.RequireRole("admin", adminHandler.GetAllUsers)))
}

// healthHandler provides a simple health check endpoint
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "message": "Backend server is running"}`))
}

// serveStaticFiles serves the frontend HTML file and static assets
func (s *Server) serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	// Only serve GET requests for static files
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Serve the main HTML file for root path
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "frontend/index.html")
		return
	}

	// For other paths, try to serve static files
	filePath := "web" + r.URL.Path
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, serve the main HTML file (for SPA routing)
		http.ServeFile(w, r, "frontend/index.html")
		return
	}

	// Serve the requested file
	http.ServeFile(w, r, filePath)
}

// corsMiddleware handles Cross-Origin Resource Sharing for frontend communication
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers to allow frontend communication
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
