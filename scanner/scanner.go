package scanner

import (
	"database/sql"
	"fmt"
)

// Scanner used for scan target
type Scanner interface {
	// ScanTargets return the slice of scan target
	ScanTargets(columns []string) []any
}

// ScannerPtr force the pointer
type ScannerPtr[T any] interface {
	*T
	Scanner
}

// ScanStruct scan single row
func ScanStruct[T any, P ScannerPtr[T]](rows *sql.Rows) (T, error) {
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

// ScanStructs Scan multiple structs
func ScanStructs[T any, P ScannerPtr[T]](rows *sql.Rows) ([]*T, error) {
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

// QueryStruct single query
func QueryStruct[T any, P ScannerPtr[T]](db *sql.DB, query string, args ...any) (T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		var zero T
		return zero, err
	}
	defer rows.Close()

	return ScanStruct[T, P](rows)
}

// QueryStructs query struct array
func QueryStructs[T any, P ScannerPtr[T]](db *sql.DB, query string, args ...any) ([]*T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return ScanStructs[T, P](rows)
}

// ScanMap Helper function, create needed mapping
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
