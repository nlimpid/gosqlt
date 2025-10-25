// Package scanner provides type-safe, generic utilities for scanning SQL query results into Go structs.
//
// This package simplifies the process of mapping database rows to struct fields by leveraging
// Go generics and the Scanner interface. It eliminates boilerplate code typically required
// when working with database/sql.
//
// # Basic Usage
//
// To use this package, implement the Scanner interface for your struct type:
//
//	type User struct {
//	    ID   int64
//	    Name string
//	    Age  int
//	}
//
//	func (u *User) ScanTargets(columns []string) []any {
//	    return scanner.ScanMap(columns, map[string]any{
//	        "id":   &u.ID,
//	        "name": &u.Name,
//	        "age":  &u.Age,
//	    })
//	}
//
// # Query Single Row
//
// Use QueryStruct to fetch and scan a single row:
//
//	user, err := scanner.QueryStruct[User](ctx, db,
//	    "SELECT id, name, age FROM users WHERE id = ?", 1)
//	if err != nil {
//	    return err
//	}
//
// # Query Multiple Rows
//
// Use QueryStructs to fetch and scan multiple rows:
//
//	users, err := scanner.QueryStructs[User](ctx, db,
//	    "SELECT id, name, age FROM users", nil)
//	if err != nil {
//	    return err
//	}
//
// # Performance Optimization
//
// When querying large result sets, pre-allocate slice capacity:
//
//	users, err := scanner.QueryStructs[User](ctx, db,
//	    "SELECT id, name, age FROM users", nil,
//	    scanner.WithExpectedSize(1000))
//
// # Custom Scanners
//
// The Scanner interface gives you full control over the scan destinations. Most callers can
// rely on ScanMap, but you can build your own ScanTargets implementation when you need to
// handle virtual or computed fields.
//
//	type AuditLog struct {
//	    Payload []byte
//	}
//
//	func (a *AuditLog) ScanTargets(columns []string) []any {
//	    targets := make([]any, len(columns))
//	    for i, col := range columns {
//	        switch col {
//	        case "payload":
//	            targets[i] = &a.Payload
//	        default:
//	            var discard any
//	            targets[i] = &discard
//	        }
//	    }
//	    return targets
//	}
//
// # Query Options
//
// QueryOption hooks provide light-weight tuneables without introducing a builder-style API.
// Use WithExpectedSize when you know the approximate row count to reduce slice reallocations.
//
// # Low-Level API
//
// For more control, use ScanStruct and ScanStructs directly with sql.Rows:
//
//	rows, err := db.QueryContext(ctx, "SELECT * FROM users")
//	if err != nil {
//	    return err
//	}
//	defer rows.Close()
//
//	users, err := scanner.ScanStructs[User](rows)
package scanner
