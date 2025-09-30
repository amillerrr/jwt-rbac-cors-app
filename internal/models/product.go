package models

import (
	"time"

	"github.com/amillerrr/jwt-rbac-cors-app/internal/database"
)

// Product represents a product in the system
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	UserID      *int      `json:"user_id"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductRepository handles database operations for products
type ProductRepository struct {
	db database.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db database.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// GetAll retrieves all active products
func (r *ProductRepository) GetAll() ([]Product, error) {
	query := `
		SELECT id, name, description, price, user_id, is_active, created_at, updated_at
		FROM products 
		WHERE is_active = true 
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description,
			&product.Price, &product.UserID, &product.IsActive, 
			&product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

// GetByID retrieves a specific product by ID
func (r *ProductRepository) GetByID(id int) (*Product, error) {
	product := &Product{}
	query := `
		SELECT id, name, description, price, user_id, is_active, created_at, updated_at
		FROM products 
		WHERE id = $1 AND is_active = true`

	err := r.db.QueryRow(query, id).Scan(
		&product.ID, &product.Name, &product.Description,
		&product.Price, &product.UserID, &product.IsActive, 
		&product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return product, nil
}

// GetByUserID retrieves all products created by a specific user
func (r *ProductRepository) GetByUserID(userID int) ([]Product, error) {
	query := `
		SELECT id, name, description, price, user_id, is_active, created_at, updated_at
		FROM products 
		WHERE user_id = $1 AND is_active = true 
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description,
			&product.Price, &product.UserID, &product.IsActive, 
			&product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}
