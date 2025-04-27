package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"brokerapp/pkg/circuitbreaker"

	"github.com/sony/gobreaker"
)

type MySQL struct {
	db *sql.DB
	cb *circuitbreaker.CircuitBreaker
}

func NewMySQL(dsn string) (*MySQL, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Create circuit breaker for database operations
	cb := circuitbreaker.New("mysql-db",
		circuitbreaker.WithMaxRequests(3),
		circuitbreaker.WithInterval(30*time.Second),
		circuitbreaker.WithTimeout(10*time.Second),
		circuitbreaker.WithReadyToTrip(func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 3
		}),
		circuitbreaker.WithOnStateChange(func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit breaker %s state changed from %s to %s\n", name, from, to)
		}),
	)

	return &MySQL{
		db: db,
		cb: cb,
	}, nil
}

// Query executes a query that returns rows
func (m *MySQL) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	var rows *sql.Rows
	var err error

	_, err = m.cb.ExecuteWithBreaker(ctx, func() (interface{}, error) {
		rows, err = m.db.QueryContext(ctx, query, args...)
		return rows, err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return rows, nil
}

// QueryRow executes a query that is expected to return at most one row
func (m *MySQL) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	var row *sql.Row

	_, err := m.cb.ExecuteWithBreaker(ctx, func() (interface{}, error) {
		row = m.db.QueryRowContext(ctx, query, args...)
		return row, nil
	})

	if err != nil {
		// Return a row that will always return an error
		return &sql.Row{}
	}

	return row
}

// Exec executes a query without returning any rows
func (m *MySQL) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	_, err = m.cb.ExecuteWithBreaker(ctx, func() (interface{}, error) {
		result, err = m.db.ExecContext(ctx, query, args...)
		return result, err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return result, nil
}

// Close closes the database connection
func (m *MySQL) Close() error {
	return m.db.Close()
}
