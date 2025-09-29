package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/amillerrr/jwt-rbac-cors-app/internal/auth"
	"github.com/amillerrr/jwt-rbac-cors-app/internal/models"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	productRepo *models.ProductRepository
}

// NewProductHandler creates a new product handler
func NewProductHandler(db *sql.DB) *ProductHandler {
	return &ProductHandler{
		productRepo: models.NewProductRepository(db),
	}
}

// GetProducts returns all products (protected endpoint)
func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get products from database
	products, err := h.productRepo.GetAll()
	if err != nil {
		http.Error(w, "Failed to retrieve products", http.StatusInternalServerError)
		return
	}

	// Return products as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

// GetProduct returns a specific product by ID
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract product ID from URL path
	// This is a simple approach - in production you'd use a router like gorilla/mux
	path := strings.TrimPrefix(r.URL.Path, "/products/")
	if path == "" {
		http.Error(w, "Product ID required", http.StatusBadRequest)
		return
	}

	productID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Get product from database
	product, err := h.productRepo.GetByID(productID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve product", http.StatusInternalServerError)
		return
	}

	// Return product as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// GetMyProducts returns products created by the current user
func (h *ProductHandler) GetMyProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by authentication middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusInternalServerError)
		return
	}

	// Get user's products from database
	products, err := h.productRepo.GetByUserID(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve products", http.StatusInternalServerError)
		return
	}

	// Return products as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// CreateProduct creates a new product (authenticated users only)
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// This would be implemented in a more complete application
	// For now, we'll return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Product creation endpoint - implementation pending",
		"status":  "placeholder",
	})
}

// UpdateProduct updates an existing product
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// This would be implemented in a more complete application
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Product update endpoint - implementation pending",
		"status":  "placeholder",
	})
}

// DeleteProduct deletes a product (soft delete by setting is_active = false)
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// This would be implemented in a more complete application
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Product deletion endpoint - implementation pending",
		"status":  "placeholder",
	})
}
