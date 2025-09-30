package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/amillerrr/jwt-rbac-cors-app/internal/monitoring"
)

type InstrumentedDB struct {
	*sql.DB
	metrics *monitoring.Metrics
}

func NewInstrumentedDB(db *sql.DB, metrics *monitoring.Metrics) *InstrumentedDB {
	return &InstrumentedDB{
		DB:      db,
		metrics: metrics,
	}
}

func (idb *InstrumentedDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		idb.metrics.DBQueryDuration.WithLabelValues("query_row").Observe(duration.Seconds())
		idb.metrics.DBQueriesTotal.WithLabelValues("query_row", "success").Inc()
	}()
	return idb.DB.QueryRowContext(ctx, query, args...)
}

func (idb *InstrumentedDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return idb.QueryRowContext(context.Background(), query, args...)
}

func (idb *InstrumentedDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return idb.QueryContext(context.Background(), query, args...)
}

func (idb *InstrumentedDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := idb.DB.QueryContext(ctx, query, args...)
	
	duration := time.Since(start)
	idb.metrics.DBQueryDuration.WithLabelValues("query").Observe(duration.Seconds())
	
	status := "success"
	if err != nil {
		status = "error"
	}
	idb.metrics.DBQueriesTotal.WithLabelValues("query", status).Inc()
	
	return rows, err
}

func (idb *InstrumentedDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return idb.ExecContext(context.Background(), query, args...)
}

func (idb *InstrumentedDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := idb.DB.ExecContext(ctx, query, args...)
	
	duration := time.Since(start)
	idb.metrics.DBQueryDuration.WithLabelValues("exec").Observe(duration.Seconds())
	
	status := "success"
	if err != nil {
		status = "error"
	}
	idb.metrics.DBQueriesTotal.WithLabelValues("exec", status).Inc()
	
	return result, err
}
