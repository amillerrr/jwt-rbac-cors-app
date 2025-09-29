package database

import (
	"context"
	"database/sql"
)

// DB defines the interface for database operations needed by monitoring
// This interface allows us to work with *sql.DB without tight coupling
type DB interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Stats() sql.DBStats
	Ping() error
	Close() error
}

// Ensure *sql.DB implements our DB interface at compile time
var _ DB = (*sql.DB)(nil)
