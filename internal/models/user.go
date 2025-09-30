package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/amillerrr/jwt-rbac-cors-app/internal/database"
)

// User represents a user in the system
type User struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"-"` // Never send password hash in JSON
	EmailVerified bool       `json:"email_verified"`
	IsActive      bool       `json:"is_active"`
	LastLogin     *time.Time `json:"last_login,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Roles         []string   `json:"roles,omitempty"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents successful login response
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// CreateUserRequest represents user registration data
type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByEmail retrieves a user by email address
func (r *UserRepository) GetByEmail(email string) (*User, error) {
	user := &User{}
	query := `
		SELECT u.id, u.name, u.email, u.password_hash, u.email_verified, 
		       u.is_active, u.last_login, u.created_at, u.updated_at
		FROM users u 
		WHERE u.email = $1 AND u.is_active = true`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.EmailVerified, &user.IsActive, &user.LastLogin, 
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Get user roles
	roles, err := r.getUserRoles(user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	return user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id int) (*User, error) {
	user := &User{}
	query := `
		SELECT u.id, u.name, u.email, u.password_hash, u.email_verified, 
		       u.is_active, u.last_login, u.created_at, u.updated_at
		FROM users u 
		WHERE u.id = $1 AND u.is_active = true`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.EmailVerified, &user.IsActive, &user.LastLogin, 
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Get user roles
	roles, err := r.getUserRoles(user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	return user, nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(userID int) error {
	query := "UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = $1"
	_, err := r.db.Exec(query, userID)
	return err
}

// getUserRoles retrieves all roles for a specific user
func (r *UserRepository) getUserRoles(userID int) ([]string, error) {
	query := `
		SELECT r.name 
		FROM roles r 
		JOIN user_roles ur ON r.id = ur.role_id 
		WHERE ur.user_id = $1`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// HasRole checks if a user has a specific role
func (u *User) HasRole(role string) bool {
	for _, userRole := range u.Roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// Create creates a new user in the database
func (r *UserRepository) Create(user *User) error {
	// Start a transaction for creating user and assigning default role
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Insert the user
	query := `
		INSERT INTO users (name, email, password_hash, email_verified, is_active) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, created_at, updated_at`

	err = tx.QueryRow(query, user.Name, user.Email, user.PasswordHash, user.EmailVerified, user.IsActive).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Assign default "user" role
	roleQuery := `
		INSERT INTO user_roles (user_id, role_id) 
		SELECT $1, id FROM roles WHERE name = 'user'`
	
	_, err = tx.Exec(roleQuery, user.ID)
	if err != nil {
		return fmt.Errorf("failed to assign default role: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Set the default role in the user object
	user.Roles = []string{"user"}
	
	return nil
}

// EmailExists checks if an email address is already registered
func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE email = $1"
	err := r.db.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
