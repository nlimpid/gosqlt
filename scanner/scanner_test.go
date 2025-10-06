package scanner

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/slog"
)

type ScannerTestDuckdbSuite struct {
	suite.Suite
	db *sql.DB
}

func (s *ScannerTestDuckdbSuite) SetupSuite() {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		s.T().Fatal(err)
	}
	s.db = db

}

func Test_ScannerTestDuckdbSuite(t *testing.T) {
	suite.Run(t, new(ScannerTestDuckdbSuite))
}

type Foo struct {
	Id        int64
	Name      string
	ClassTime time.Time
	Deposit   float32
}

func (f *Foo) ScanTargets(columns []string) []any {
	// 使用辅助函数简化实现
	return ScanMap(columns, map[string]any{
		"id":         &f.Id,
		"name":       &f.Name,
		"class_time": &f.ClassTime,
		"deposit":    &f.Deposit,
	})
}

func (s *ScannerTestDuckdbSuite) Test_QueryStruct() {
	result, err := QueryStruct[Foo](s.db, `SELECT 1 as id, 'foo' as name, 2 as age, '100.00003' as deposit`)
	if err != nil {
		s.T().Fatal(err)
	}
	slog.Info("result %+v", result)
}

func TestQueryStructs_NoCache(t *testing.T) {
	db, _ := sql.Open("duckdb", "")
	defer db.Close()

	query := `
          SELECT
              i as id,
              'name_' || i as name,
              i % 100 as age
          FROM generate_series(1, 1000000) as t(i)
      `
	start := time.Now()
	_, err := QueryStructs[Foo](db, query)
	if err != nil {
		t.Fatal(err)
	}

	slog.Info("took", "delta", time.Since(start))
}

// Benchmark 测试
func BenchmarkExample(b *testing.B) {
	db, _ := sql.Open("duckdb", "")
	defer db.Close()
	s := new(ScannerTestDuckdbSuite)
	s.SetT(&testing.T{})
	s.SetupSuite()
	b.StartTimer()
	query := `
          SELECT
              i as id,
              'name_' || i as name,
              i % 100 as age
          FROM generate_series(1, 1000000) as t(i)
      `
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := QueryStructs[Foo](db, query)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}
