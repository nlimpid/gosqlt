# gosqlt

A type-safe, generic scanner library for Go that simplifies scanning SQL query results into structs.

## Quick Start

### 1. Define your struct and implement the Scanner interface

```go
type User struct {
    ID   int64
    Name string
    Age  int
}

func (u *User) ScanTargets(columns []string) []any {
    return scanner.ScanMap(columns, map[string]any{
        "id":   &u.ID,
        "name": &u.Name,
        "age":  &u.Age,
    })
}
```

### 2. Query a single row

```go
user, err := scanner.QueryStruct[User](context.Background(), db,
    "SELECT id, name, age FROM users WHERE id = ?", 1)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User: %+v\n", user)
```

### 3. Query multiple rows

```go
users, err := scanner.QueryStructs[User](context.Background(), db,
    "SELECT id, name, age FROM users", nil)
if err != nil {
    log.Fatal(err)
}
for _, user := range users {
    fmt.Printf("User: %+v\n", *user)
}
```

### 4. Use query options for better performance

```go
// Pre-allocate slice capacity when you know the expected result size
users, err := scanner.QueryStructs[User](context.Background(), db,
    "SELECT id, name, age FROM users", nil,
    scanner.WithExpectedSize(1000))
```
