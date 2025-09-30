package database

import (
	"context"
	"database/sql"
)

// DB defines the interface for database operations
type DB interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Begin() (*sql.Tx, error)
	Stats() sql.DBStats
	Ping() error
	Close() error
}

// Ensure *sql.DB implements our DB interface at compile time
var _ DB = (*sql.DB)(nil)
