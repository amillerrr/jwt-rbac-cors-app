package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/amillerrr/jwt-rbac-cors-app/internal/auth"
)

// AdminHandler handles admin-only HTTP requests
type AdminHandler struct {
	db *sql.DB
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{
		db: db,
	}
}

// GetAdminData returns admin-only information
func (h *AdminHandler) GetAdminData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user information from context
	userEmail, ok := auth.GetUserEmailFromContext(r.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusInternalServerError)
		return
	}

	userRoles, _ := auth.GetUserRolesFromContext(r.Context())

	// Prepare admin response
	response := map[string]interface{}{
		"message":     "This is admin-only content!",
		"user":        userEmail,
		"roles":       userRoles,
		"admin_info": map[string]interface{}{
			"total_users":    h.getTotalUsers(),
			"total_products": h.getTotalProducts(),
			"system_status":  "operational",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSystemStats returns system statistics (admin only)
func (h *AdminHandler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := map[string]interface{}{
		"users": map[string]interface{}{
			"total":         h.getTotalUsers(),
			"active":        h.getActiveUsers(),
			"verified":      h.getVerifiedUsers(),
			"recent_logins": h.getRecentLogins(),
		},
		"products": map[string]interface{}{
			"total":  h.getTotalProducts(),
			"active": h.getActiveProducts(),
		},
		"system": map[string]interface{}{
			"database_status": h.checkDatabaseHealth(),
			"uptime":         "N/A", // Would be calculated in a real system
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetAllUsers returns all users (admin only)
func (h *AdminHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `
		SELECT id, name, email, email_verified, is_active, created_at, last_login
		FROM users 
		ORDER BY created_at DESC`

	rows, err := h.db.Query(query)
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var name, email string
		var emailVerified, isActive bool
		var createdAt string
		var lastLogin sql.NullString

		err := rows.Scan(&id, &name, &email, &emailVerified, &isActive, &createdAt, &lastLogin)
		if err != nil {
			continue // Skip problematic rows
		}

		user := map[string]interface{}{
			"id":             id,
			"name":           name,
			"email":          email,
			"email_verified": emailVerified,
			"is_active":      isActive,
			"created_at":     createdAt,
			"last_login":     nil,
		}

		if lastLogin.Valid {
			user["last_login"] = lastLogin.String
		}

		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Helper functions for gathering statistics

func (h *AdminHandler) getTotalUsers() int {
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (h *AdminHandler) getActiveUsers() int {
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = true").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (h *AdminHandler) getVerifiedUsers() int {
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM users WHERE email_verified = true").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (h *AdminHandler) getRecentLogins() int {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE last_login > NOW() - INTERVAL '24 hours'"
	err := h.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (h *AdminHandler) getTotalProducts() int {
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (h *AdminHandler) getActiveProducts() int {
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM products WHERE is_active = true").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (h *AdminHandler) checkDatabaseHealth() string {
	err := h.db.Ping()
	if err != nil {
		return "unhealthy"
	}
	return "healthy"
}
