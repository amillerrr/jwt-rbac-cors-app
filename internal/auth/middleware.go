package auth

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// UserEmailKey is the context key for user email
	UserEmailKey ContextKey = "user_email"
	// UserRolesKey is the context key for user roles
	UserRolesKey ContextKey = "user_roles"
)

// Middleware provides authentication and authorization middleware
type Middleware struct {
	jwtService *JWTService
}

// NewMiddleware creates a new authentication middleware
func NewMiddleware(jwtService *JWTService) *Middleware {
	return &Middleware{
		jwtService: jwtService,
	}
}

// RequireAuth ensures the request has a valid JWT token
func (m *Middleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
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

		tokenString := parts[1]

		// Validate the token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user information to request context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)

		// Call next handler with updated context
		next(w, r.WithContext(ctx))
	}
}

// RequireRole ensures the user has a specific role
func (m *Middleware) RequireRole(role string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			// Get user roles from context
			rolesInterface := r.Context().Value(UserRolesKey)
			roles, ok := rolesInterface.([]string)
			if !ok {
				http.Error(w, "Unable to verify user roles", http.StatusInternalServerError)
				return
			}

			// Check if user has the required role
			hasRole := false
			for _, userRole := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// User has required role, proceed
			next(w, r)
		})
	}
}

// RequireAnyRole ensures the user has at least one of the specified roles
func (m *Middleware) RequireAnyRole(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			// Get user roles from context
			rolesInterface := r.Context().Value(UserRolesKey)
			roles, ok := rolesInterface.([]string)
			if !ok {
				http.Error(w, "Unable to verify user roles", http.StatusInternalServerError)
				return
			}

			// Check if user has any of the allowed roles
			hasRole := false
			for _, userRole := range roles {
				for _, allowedRole := range allowedRoles {
					if userRole == allowedRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// User has at least one allowed role, proceed
			next(w, r)
		})
	}
}

// GetUserIDFromContext extracts the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}

// GetUserEmailFromContext extracts the user email from the request context
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

// GetUserRolesFromContext extracts the user roles from the request context
func GetUserRolesFromContext(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(UserRolesKey).([]string)
	return roles, ok
}

// Legacy header-based approach for backward compatibility
// This is the approach used in the original code - we keep it for comparison
func (m *Middleware) RequireAuthLegacy(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := m.jwtService.ValidateToken(bearerToken[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user info to request headers (legacy approach)
		r.Header.Set("X-User-ID", strconv.Itoa(claims.UserID))
		r.Header.Set("X-User-Email", claims.Email)
		r.Header.Set("X-User-Roles", strings.Join(claims.Roles, ","))

		next(w, r)
	}
}
