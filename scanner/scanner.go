package scanner

import (
	"context"
	"database/sql"
	"fmt"
)

// Scanner describes a type that knows how to turn a set of database column
// names into destinations that the standard library's sql.Rows can write into.
// See ScanTargets for the detailed contract.
type Scanner interface {
	// ScanTargets returns a slice of pointers matching the provided columns.
	// Each entry must be safe to pass to (*sql.Rows).Scan in the same order.
	ScanTargets(columns []string) []any
}

// Ptr is a generic type constraint requiring a pointer to T that also implements
// Scanner. It lets ScanStruct and friends control the creation of new values
// while still letting the user provide custom ScanTargets logic.
type Ptr[T any] interface {
	*T
	Scanner
}

// QueryOption configures query behavior.
type QueryOption func(*queryConfig)

type queryConfig struct {
	expectedSize int
}

// WithExpectedSize pre-allocates slice capacity for better performance. It is
// primarily used with ScanStructs and QueryStructs when you know the expected
// number of rows ahead of time.
func WithExpectedSize(size int) QueryOption {
	return func(c *queryConfig) {
		c.expectedSize = size
	}
}

// ScanStruct reads the first row from rows and decodes it into a new struct
// value. It stops after the first row and returns sql.ErrNoRows when the result
// set is empty.
func ScanStruct[T any, P Ptr[T]](rows *sql.Rows) (*T, error) {
	var result T

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Use pointer to get scan targets
	ptr := P(&result)
	targets := ptr.ScanTargets(columns)

	if err := rows.Scan(targets...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	return &result, nil
}

// ScanStructs consumes rows and returns one pointer per row in the order they
// are produced by the driver. Combine it with WithExpectedSize to avoid slice
// resizing during large iterations.
func ScanStructs[T any, P Ptr[T]](rows *sql.Rows, opts ...QueryOption) ([]*T, error) {
	cfg := &queryConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	results := make([]*T, 0, cfg.expectedSize)
	for rows.Next() {
		var result T
		ptr := P(&result)
		targets := ptr.ScanTargets(columns)
		if err := rows.Scan(targets...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}

// QueryStruct runs query against db with the provided args, then delegates to
// ScanStruct. The underlying rows cursor is closed automatically.
func QueryStruct[T any, P Ptr[T]](ctx context.Context, db *sql.DB, query string, args ...any) (*T, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return ScanStruct[T, P](rows)
}

// QueryStructs runs query against db with the supplied args slice, applying the
// given QueryOptions before delegating to ScanStructs. The returned rows cursor
// is closed automatically.
func QueryStructs[T any, P Ptr[T]](ctx context.Context, db *sql.DB, query string, args []any, opts ...QueryOption) ([]*T, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return ScanStructs[T, P](rows, opts...)
}

// ScanMap creates a ScanTargets-compatible slice from a column-to-field map.
// Columns not present in mapping receive a throwaway placeholder pointer so the
// caller can ignore unexpected projections safely.
func ScanMap(columns []string, mapping map[string]any) []any {
	targets := make([]any, len(columns))
	for i, col := range columns {
		if target, ok := mapping[col]; ok {
			targets[i] = target
		} else {
			var placeholder any
			targets[i] = &placeholder
		}
	}
	return targets
}
