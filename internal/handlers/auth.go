package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"log/slog"

	"github.com/amillerrr/jwt-rbac-cors-app/internal/auth"
	"github.com/amillerrr/jwt-rbac-cors-app/internal/models"
	"github.com/amillerrr/jwt-rbac-cors-app/pkg/crypto"
	"github.com/amillerrr/jwt-rbac-cors-app/pkg/validator"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userRepo   *models.UserRepository
	jwtService *auth.JWTService
	middleware *auth.Middleware
	logger     *slog.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *sql.DB, jwtSecret string, logger *slog.Logger) *AuthHandler {
	jwtService := auth.NewJWTService(jwtSecret)
	return &AuthHandler{
		userRepo:   models.NewUserRepository(db),
		jwtService: jwtService,
		middleware: auth.NewMiddleware(jwtService),
		logger: logger,
	}
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse login request
	var loginReq models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate input
	if loginReq.Email == "" || loginReq.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Find user by email
	user, err := h.userRepo.GetByEmail(loginReq.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Verify password
	if !crypto.CheckPasswordHash(loginReq.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Update last login timestamp
	if err := h.userRepo.UpdateLastLogin(user.ID); err != nil {
		h.logger.Error("Failed to update last login timestamp",
			slog.String("error", err.Error()),
			slog.String("handler", "Login"),
		)
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode JSON response",
			slog.String("error", err.Error()),
			slog.String("handler", "Login"),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse registration request
	var registerReq models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate input
	validationErrors := validator.ValidateUserRegistration(registerReq.Name, registerReq.Email, registerReq.Password)
	if validationErrors.HasErrors() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Validation failed",
			"details": validationErrors,
		})
		return
	}

	// Check if email already exists
	emailExists, err := h.userRepo.EmailExists(registerReq.Email)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if emailExists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Email already registered",
			"details": []validator.ValidationError{
				{Field: "email", Message: "An account with this email already exists"},
			},
		})
		return
	}

	// Hash the password
	passwordHash, err := crypto.HashPassword(registerReq.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user object
	user := &models.User{
		Name:          strings.TrimSpace(registerReq.Name),
		Email:         strings.ToLower(strings.TrimSpace(registerReq.Email)),
		PasswordHash:  passwordHash,
		EmailVerified: false, // In production, you'd send a verification email
		IsActive:      true,
	}

	// Save user to database
	if err := h.userRepo.Create(user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT token for immediate login
	token, err := h.jwtService.GenerateToken(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Prepare response (same as login response)
	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// RefreshToken handles token refresh requests
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract current token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Generate new token
	newToken, err := h.jwtService.RefreshToken(parts[1])
	if err != nil {
		http.Error(w, "Cannot refresh token", http.StatusUnauthorized)
		return
	}

	// Send new token
	response := map[string]string{"token": newToken}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusInternalServerError)
		return
	}

	// Fetch user from database
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return user profile
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// RequireAuth wraps handlers that require authentication
func (h *AuthHandler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return h.middleware.RequireAuth(next)
}

// RequireRole wraps handlers that require a specific role
func (h *AuthHandler) RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return h.middleware.RequireRole(role)(next)
}

// RequireAnyRole wraps handlers that require any of the specified roles
func (h *AuthHandler) RequireAnyRole(next http.HandlerFunc, roles ...string) http.HandlerFunc {
	return h.middleware.RequireAnyRole(roles...)(next)
}
