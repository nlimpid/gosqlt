package scanner

import (
	"context"
	"database/sql"
	"fmt"
)

// Scanner is an interface for scanning database rows into struct fields.
type Scanner interface {
	// ScanTargets returns a slice of pointers to scan targets for the given columns.
	ScanTargets(columns []string) []any
}

// Ptr is an interface constraint that requires a pointer to T implementing Scanner.
// It's used as a hook for ScanStruct to ensure type safety.
type Ptr[T any] interface {
	*T
	Scanner
}

// ScanStruct scans a single row from the result set into a struct.
func ScanStruct[T any, P Ptr[T]](rows *sql.Rows) (T, error) {
	var result T

	if !rows.Next() {
		return result, sql.ErrNoRows
	}

	columns, err := rows.Columns()
	if err != nil {
		return result, fmt.Errorf("failed to get columns: %w", err)
	}

	// Use pointer to get scan targets
	ptr := P(&result)
	targets := ptr.ScanTargets(columns)

	if err := rows.Scan(targets...); err != nil {
		return result, fmt.Errorf("failed to scan row: %w", err)
	}

	return result, nil
}

// ScanStructs scans multiple rows from the result set into a slice of struct pointers.
func ScanStructs[T any, P Ptr[T]](rows *sql.Rows) ([]*T, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	// TODO: maybe can fix it
	results := make([]*T, 0)
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

// QueryStruct executes a query and scans a single row into a struct.
func QueryStruct[T any, P Ptr[T]](ctx context.Context, db *sql.DB, query string, args ...any) (T, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		var zero T
		return zero, err
	}
	defer rows.Close()

	return ScanStruct[T, P](rows)
}

// QueryStructs executes a query and scans multiple rows into a slice of struct pointers.
func QueryStructs[T any, P Ptr[T]](ctx context.Context, db *sql.DB, query string, args ...any) ([]*T, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return ScanStructs[T, P](rows)
}

// ScanMap is a helper function that creates a slice of scan targets based on a column-to-field mapping.
// Columns not found in the mapping will use a placeholder.
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
