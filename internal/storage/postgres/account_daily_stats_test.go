package postgres_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kraconst/GO_server/internal/domain"
	"github.com/kraconst/GO_server/internal/storage/postgres"
)

type mockDriver struct {
	mu           sync.Mutex
	queryHandler func(query string, args []driver.Value) (driver.Rows, error)
	execHandler  func(query string, args []driver.Value) (driver.Result, error)
}

var (
	driverInstance = &mockDriver{}
	driverOnce     sync.Once
)

func init() {
	driverOnce.Do(func() {
		sql.Register("mock_pg_stats", driverInstance)
	})
}

func (d *mockDriver) Open(name string) (driver.Conn, error) {
	return &mockConn{driver: d}, nil
}

type mockConn struct {
	driver *mockDriver
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return &mockStmt{conn: c, query: query}, nil
}
func (c *mockConn) Close() error { return nil }
func (c *mockConn) Begin() (driver.Tx, error) {
	return &mockTx{}, nil
}

type mockTx struct{}

func (t *mockTx) Commit() error   { return nil }
func (t *mockTx) Rollback() error { return nil }

type mockStmt struct {
	conn  *mockConn
	query string
}

func (s *mockStmt) Close() error  { return nil }
func (s *mockStmt) NumInput() int { return -1 }

func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	s.conn.driver.mu.Lock()
	defer s.conn.driver.mu.Unlock()
	if s.conn.driver.execHandler != nil {
		return s.conn.driver.execHandler(s.query, args)
	}
	return mockResult{rowsAffected: 1}, nil
}

func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	s.conn.driver.mu.Lock()
	defer s.conn.driver.mu.Unlock()
	if s.conn.driver.queryHandler != nil {
		return s.conn.driver.queryHandler(s.query, args)
	}
	return &mockRows{}, nil
}

type mockResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r mockResult) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r mockResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type mockRows struct {
	cols []string
	vals [][]driver.Value
	idx  int
}

func (r *mockRows) Columns() []string { return r.cols }
func (r *mockRows) Close() error      { return nil }
func (r *mockRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.vals) {
		return io.EOF
	}
	for i, v := range r.vals[r.idx] {
		dest[i] = v
	}
	r.idx++
	return nil
}

func setupMockDB(t *testing.T) (*sqlx.DB, *mockDriver) {
	t.Helper()
	db, err := sql.Open("mock_pg_stats", "test")
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "postgres")
	return sqlxDB, driverInstance
}

func TestAccountDailyStatsStorage_Upsert(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	storage := postgres.NewAccountDailyStatsStorage(db)
	ctx := context.Background()

	var capturedQuery string
	var capturedArgs []driver.Value

	mock.execHandler = func(query string, args []driver.Value) (driver.Result, error) {
		capturedQuery = query
		capturedArgs = args
		return mockResult{rowsAffected: 1}, nil
	}

	day := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	err := storage.Upsert(ctx, domain.AccountDailyStats{
		AccountID:       42,
		Day:             day,
		MessagesIn:      15,
		MessagesOut:     10,
		ActiveChats:     5,
		NewContacts:     2,
		AvgResponseTime: 45 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cleanQ := strings.ToLower(strings.Join(strings.Fields(capturedQuery), " "))
	if !strings.Contains(cleanQ, "insert into account_daily_stats") {
		t.Errorf("query missing 'insert into account_daily_stats': %s", capturedQuery)
	}
	if !strings.Contains(cleanQ, "on conflict (account_id, day)") {
		t.Errorf("query missing 'on conflict (account_id, day)': %s", capturedQuery)
	}
	if len(capturedArgs) != 7 {
		t.Fatalf("expected 7 args, got %d", len(capturedArgs))
	}
	if capturedArgs[0] != int64(42) || capturedArgs[1] != "2026-11-10" {
		t.Errorf("unexpected args: %#v", capturedArgs)
	}
}

func TestAccountDailyStatsStorage_IncrementCounters(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	storage := postgres.NewAccountDailyStatsStorage(db)
	ctx := context.Background()

	var capturedQuery string
	var capturedArgs []driver.Value

	mock.execHandler = func(query string, args []driver.Value) (driver.Result, error) {
		capturedQuery = query
		capturedArgs = args
		return mockResult{rowsAffected: 1}, nil
	}

	day := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	err := storage.IncrementCounters(ctx, 42, day, 3, 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cleanQ := strings.ToLower(strings.Join(strings.Fields(capturedQuery), " "))
	if !strings.Contains(cleanQ, "on conflict (account_id, day) do update set") {
		t.Errorf("query missing atomic increment conflict update: %s", capturedQuery)
	}
	if len(capturedArgs) != 5 {
		t.Fatalf("expected 5 args, got %d", len(capturedArgs))
	}
}

func TestAccountDailyStatsStorage_GetDashboard(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	storage := postgres.NewAccountDailyStatsStorage(db)
	ctx := context.Background()

	mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
		return &mockRows{
			cols: []string{"messages_in", "messages_out", "active_chats", "new_contacts", "avg_response_time", "unanswered_chats"},
			vals: [][]driver.Value{{int64(100), int64(80), int64(15), int64(5), int64(3000000000), int64(0)}},
		}, nil
	}

	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)

	dash, err := storage.GetDashboard(ctx, 1, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dash.MessagesIn != 100 || dash.MessagesOut != 80 {
		t.Errorf("unexpected dashboard data: %+v", dash)
	}
}

func TestMigrationsSchemaIntegrity(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	repoRoot := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	upPath := filepath.Join(repoRoot, "migrations", "000004_create_account_daily_stats.up.sql")
	downPath := filepath.Join(repoRoot, "migrations", "000004_create_account_daily_stats.down.sql")

	upBytes, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("failed to read up migration: %v", err)
	}
	upContent := string(upBytes)

	downBytes, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("failed to read down migration: %v", err)
	}
	downContent := string(downBytes)

	if !regexp.MustCompile(`(?i)create\s+table\s+(if\s+not\s+exists\s+)?account_daily_stats`).MatchString(upContent) {
		t.Error("missing CREATE TABLE account_daily_stats")
	}

	reqCols := []string{
		`account_id\s+BIGINT\s+NOT\s+NULL\s+REFERENCES\s+accounts\(id\)\s+ON\s+DELETE\s+CASCADE`,
		`day\s+DATE\s+NOT\s+NULL`,
		`messages_in\s+INT\s+NOT\s+NULL\s+DEFAULT\s+0`,
		`messages_out\s+INT\s+NOT\s+NULL\s+DEFAULT\s+0`,
		`active_chats\s+INT\s+NOT\s+NULL\s+DEFAULT\s+0`,
		`new_contacts\s+INT\s+NOT\s+NULL\s+DEFAULT\s+0`,
		`avg_response_time\s+BIGINT\s+NOT\s+NULL\s+DEFAULT\s+0`,
	}
	for _, col := range reqCols {
		if !regexp.MustCompile(`(?i)` + col).MatchString(upContent) {
			t.Errorf("account_daily_stats missing column or constraint pattern: %s", col)
		}
	}

	if !regexp.MustCompile(`(?i)PRIMARY\s+KEY\s*\(\s*account_id\s*,\s*day\s*\)`).MatchString(upContent) {
		t.Error("missing PRIMARY KEY (account_id, day)")
	}

	if !regexp.MustCompile(`(?i)create\s+index\s+(if\s+not\s+exists\s+)?idx_account_daily_stats_day\s+on\s+account_daily_stats\s*\(\s*day\s*\)`).MatchString(upContent) {
		t.Error("missing index idx_account_daily_stats_day")
	}

	if !strings.Contains(strings.ToLower(downContent), "drop table if exists account_daily_stats") {
		t.Errorf("down migration missing DROP TABLE: %s", downContent)
	}
}
