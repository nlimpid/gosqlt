package tests

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/nlimpid/gosqlt/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ScannerTestDuckdbSuite struct {
	suite.Suite
	db *sql.DB
}

func (s *ScannerTestDuckdbSuite) SetupSuite() {
	db, err := sql.Open("duckdb", "")
	s.NoError(err)
	s.db = db
}

func (s *ScannerTestDuckdbSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

func Test_ScannerTestDuckdbSuite(t *testing.T) {
	suite.Run(t, new(ScannerTestDuckdbSuite))
}

type Foo struct {
	ID        int64
	Name      string
	ClassTime time.Time
	Deposit   float32
	Extra     string
}

func (f *Foo) ScanTargets(columns []string) []any {
	// Use helper function to simplify implementation
	return scanner.ScanMap(columns, map[string]any{
		"id":         &f.ID,
		"name":       &f.Name,
		"class_time": &f.ClassTime,
		"deposit":    &f.Deposit,
	})
}

type Bar struct {
	ID    int64
	Value string
}

func (b *Bar) ScanTargets(columns []string) []any {
	return scanner.ScanMap(columns, map[string]any{
		"id":    &b.ID,
		"value": &b.Value,
	})
}

func (s *ScannerTestDuckdbSuite) TestQueryStruct() {
	tests := []struct {
		name    string
		query   string
		args    []any
		want    Foo
		wantErr bool
	}{
		{
			name:  "basic select with all fields",
			query: "SELECT 1 as id, 'foo' as name, '100.00003' as deposit",
			args:  nil,
			want: Foo{
				ID:      1,
				Name:    "foo",
				Deposit: 100.00003,
			},
			wantErr: false,
		},
		{
			name:  "select with partial fields",
			query: "SELECT 42 as id, 'bar' as name",
			args:  nil,
			want: Foo{
				ID:   42,
				Name: "bar",
			},
			wantErr: false,
		},
		{
			name:  "select with parameters",
			query: "SELECT ? as id, ? as name, ? as deposit",
			args:  []any{10, "test", float32(50.5)},
			want: Foo{
				ID:      10,
				Name:    "test",
				Deposit: 50.5,
			},
			wantErr: false,
		},
		{
			name:    "empty result set",
			query:   "SELECT 1 as id, 'foo' as name WHERE 1=0",
			args:    nil,
			want:    Foo{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := scanner.QueryStruct[Foo](context.Background(), s.db, tt.query, tt.args...)
			if tt.wantErr {
				assert.Error(s.T(), err)
			} else {
				require.NoError(s.T(), err)
				assert.Equal(s.T(), tt.want.ID, result.ID)
				assert.Equal(s.T(), tt.want.Name, result.Name)
				if tt.want.Deposit > 0 {
					assert.InDelta(s.T(), tt.want.Deposit, result.Deposit, 0.001)
				}
			}
		})
	}
}

func (s *ScannerTestDuckdbSuite) TestQueryStructs() {
	tests := []struct {
		name      string
		query     string
		args      []any
		wantCount int
		wantFirst *Bar
		wantLast  *Bar
		wantErr   bool
	}{
		{
			name:      "multiple rows",
			query:     "SELECT i as id, 'value_' || i as value FROM generate_series(1, 5) as t(i)",
			args:      nil,
			wantCount: 5,
			wantFirst: &Bar{ID: 1, Value: "value_1"},
			wantLast:  &Bar{ID: 5, Value: "value_5"},
			wantErr:   false,
		},
		{
			name:      "empty result set",
			query:     "SELECT 1 as id, 'test' as value WHERE 1=0",
			args:      nil,
			wantCount: 0,
			wantFirst: nil,
			wantLast:  nil,
			wantErr:   false,
		},
		{
			name:      "single row",
			query:     "SELECT 100 as id, 'single' as value",
			args:      nil,
			wantCount: 1,
			wantFirst: &Bar{ID: 100, Value: "single"},
			wantLast:  &Bar{ID: 100, Value: "single"},
			wantErr:   false,
		},
		{
			name:      "with parameters",
			query:     "SELECT i as id, ? || i as value FROM generate_series(?, ?) as t(i)",
			args:      []any{"item_", 10, 12},
			wantCount: 3,
			wantFirst: &Bar{ID: 10, Value: "item_10"},
			wantLast:  &Bar{ID: 12, Value: "item_12"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			results, err := scanner.QueryStructs[Bar](context.Background(), s.db, tt.query, tt.args, scanner.WithExpectedSize(tt.wantCount))
			if tt.wantErr {
				assert.Error(s.T(), err)
			} else {
				require.NoError(s.T(), err)
				assert.Equal(s.T(), tt.wantCount, len(results))

				if tt.wantCount > 0 {
					if tt.wantFirst != nil {
						assert.Equal(s.T(), tt.wantFirst.ID, results[0].ID)
						assert.Equal(s.T(), tt.wantFirst.Value, results[0].Value)
					}
					if tt.wantLast != nil {
						assert.Equal(s.T(), tt.wantLast.ID, results[len(results)-1].ID)
						assert.Equal(s.T(), tt.wantLast.Value, results[len(results)-1].Value)
					}
				}
			}
		})
	}
}

func (s *ScannerTestDuckdbSuite) TestScanStruct() {
	tests := []struct {
		name    string
		query   string
		want    Bar
		wantErr bool
	}{
		{
			name:    "scan valid row",
			query:   "SELECT 1 as id, 'test' as value",
			want:    Bar{ID: 1, Value: "test"},
			wantErr: false,
		},
		{
			name:    "scan empty result",
			query:   "SELECT 1 as id, 'test' as value WHERE 1=0",
			want:    Bar{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			rows, err := s.db.Query(tt.query)
			require.NoError(s.T(), err)
			defer rows.Close()

			result, err := scanner.ScanStruct[Bar](rows)
			if tt.wantErr {
				assert.Error(s.T(), err)
			} else {
				require.NoError(s.T(), err)
				assert.Equal(s.T(), tt.want.ID, result.ID)
				assert.Equal(s.T(), tt.want.Value, result.Value)
			}
		})
	}
}

func (s *ScannerTestDuckdbSuite) TestScanStructs() {
	tests := []struct {
		name      string
		query     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "scan multiple rows",
			query:     "SELECT i as id, 'val' || i as value FROM generate_series(1, 10) as t(i)",
			wantCount: 10,
			wantErr:   false,
		},
		{
			name:      "scan empty result",
			query:     "SELECT 1 as id, 'test' as value WHERE 1=0",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			rows, err := s.db.Query(tt.query)
			require.NoError(s.T(), err)
			defer rows.Close()

			results, err := scanner.ScanStructs[Bar](rows)
			if tt.wantErr {
				assert.Error(s.T(), err)
			} else {
				require.NoError(s.T(), err)
				assert.Equal(s.T(), tt.wantCount, len(results))
			}
		})
	}
}

func TestScanMap(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		mapping map[string]any
		verify  func(t *testing.T, result []any)
	}{
		{
			name:    "all columns mapped",
			columns: []string{"id", "name", "value"},
			mapping: map[string]any{
				"id":    new(int64),
				"name":  new(string),
				"value": new(float64),
			},
			verify: func(t *testing.T, result []any) {
				assert.Len(t, result, 3)
				assert.IsType(t, new(int64), result[0])
				assert.IsType(t, new(string), result[1])
				assert.IsType(t, new(float64), result[2])
			},
		},
		{
			name:    "partial columns mapped",
			columns: []string{"id", "name", "unknown"},
			mapping: map[string]any{
				"id":   new(int64),
				"name": new(string),
			},
			verify: func(t *testing.T, result []any) {
				assert.Len(t, result, 3)
				assert.IsType(t, new(int64), result[0])
				assert.IsType(t, new(string), result[1])
				assert.NotNil(t, result[2])
			},
		},
		{
			name:    "no columns mapped",
			columns: []string{"unknown1", "unknown2"},
			mapping: map[string]any{
				"id": new(int64),
			},
			verify: func(t *testing.T, result []any) {
				assert.Len(t, result, 2)
				// All should be placeholders
				assert.NotNil(t, result[0])
				assert.NotNil(t, result[1])
			},
		},
		{
			name:    "empty columns",
			columns: []string{},
			mapping: map[string]any{
				"id": new(int64),
			},
			verify: func(t *testing.T, result []any) {
				assert.Len(t, result, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.ScanMap(tt.columns, tt.mapping)
			tt.verify(t, result)
		})
	}
}

func TestQueryStructs_NoCache(t *testing.T) {
	db, err := sql.Open("duckdb", "")
	require.NoError(t, err)
	defer db.Close()

	query := `
          SELECT
              i as id,
              'name_' || i as name,
              i % 100 as age
          FROM generate_series(1, 1000000) as t(i)
      `
	start := time.Now()
	results, err := scanner.QueryStructs[Foo](context.Background(), db, query, nil, scanner.WithExpectedSize(1000000))
	require.NoError(t, err)
	assert.Equal(t, 1000000, len(results))

	slog.Info("took", "delta", time.Since(start))
}

func BenchmarkQueryStructs(b *testing.B) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	query := `
          SELECT
              i as id,
              'name_' || i as name,
              i % 100 as age
          FROM generate_series(1, 1000000) as t(i)
      `
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.QueryStructs[Foo](context.Background(), db, query, nil, scanner.WithExpectedSize(1000000))
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}

func BenchmarkScanMap(b *testing.B) {
	columns := []string{"id", "name", "class_time", "deposit"}
	f := &Foo{}
	mapping := map[string]any{
		"id":         &f.ID,
		"name":       &f.Name,
		"class_time": &f.ClassTime,
		"deposit":    &f.Deposit,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scanner.ScanMap(columns, mapping)
	}
}
