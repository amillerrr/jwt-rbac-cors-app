package auth

import (
	"fmt"
	"time"

	"github.com/amillerrr/jwt-rbac-cors-app/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT token claims
type Claims struct {
	UserID int      `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secret []byte
}

// NewJWTService creates a new JWT service with the provided secret
func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret: []byte(secret),
	}
}

// GenerateToken creates a new JWT token for the given user
func (j *JWTService) GenerateToken(user *models.User) (string, error) {
	// Create the token claims
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Roles:  user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "goapp",
			Subject:   fmt.Sprintf("user_%d", user.ID),
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken parses and validates a JWT token
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Check if token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Additional validation can be added here
	// For example, checking if the user still exists and is active

	return claims, nil
}

// RefreshToken creates a new token with extended expiration (optional feature)
func (j *JWTService) RefreshToken(oldToken string) (string, error) {
	claims, err := j.ValidateToken(oldToken)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	// Check if token is not too old to refresh (e.g., within last 7 days)
	if time.Since(claims.IssuedAt.Time) > 7*24*time.Hour {
		return "", fmt.Errorf("token too old to refresh")
	}

	// Create new claims with extended expiration
	newClaims := &Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Roles:  claims.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-app",
			Subject:   claims.Subject,
		},
	}

	// Create and sign new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return token.SignedString(j.secret)
}
