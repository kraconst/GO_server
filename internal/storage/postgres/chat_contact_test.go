package postgres_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
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

// mockDriver implements a mock database/sql/driver to empirically verify queries and error handling
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
		sql.Register("mock_pg", driverInstance)
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
	db, err := sql.Open("mock_pg", "test")
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "postgres")
	return sqlxDB, driverInstance
}

// ==========================================
// 1. ContactStorage Empirical Tests
// ==========================================

func TestContactStorage_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	storage := postgres.NewContactStorage(db)
	ctx := context.Background()

	t.Run("successful insert or update with returning id", func(t *testing.T) {
		var capturedQuery string
		var capturedArgs []driver.Value

		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			capturedQuery = query
			capturedArgs = args
			return &mockRows{
				cols: []string{"id"},
				vals: [][]driver.Value{{int64(42)}},
			}, nil
		}

		c := domain.Contact{
			CompanyID:      10,
			TelegramUserID: 998877,
			Name:           "Alice Smith",
			Username:       "alicesmith",
		}

		id, err := storage.Create(ctx, c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != 42 {
			t.Errorf("expected id 42, got %d", id)
		}

		cleanQ := strings.ToLower(strings.Join(strings.Fields(capturedQuery), " "))
		if !strings.Contains(cleanQ, "insert into contacts") {
			t.Errorf("query missing 'insert into contacts': %s", capturedQuery)
		}
		if !strings.Contains(cleanQ, "on conflict (company_id, telegram_user_id)") {
			t.Errorf("query missing 'on conflict (company_id, telegram_user_id)': %s", capturedQuery)
		}
		if !strings.Contains(cleanQ, "do update set name = excluded.name, username = excluded.username") {
			t.Errorf("query missing 'do update set': %s", capturedQuery)
		}
		if !strings.Contains(cleanQ, "returning id") {
			t.Errorf("query missing 'returning id': %s", capturedQuery)
		}

		if len(capturedArgs) != 4 {
			t.Fatalf("expected 4 args, got %d", len(capturedArgs))
		}
		if capturedArgs[0] != int64(10) || capturedArgs[1] != int64(998877) ||
			capturedArgs[2] != "Alice Smith" || capturedArgs[3] != "alicesmith" {
			t.Errorf("unexpected args: %#v", capturedArgs)
		}
	})

	t.Run("query error propagation", func(t *testing.T) {
		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			return nil, errors.New("db connection failure")
		}

		_, err := storage.Create(ctx, domain.Contact{CompanyID: 1, TelegramUserID: 1, Name: "A"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "upsert contact") {
			t.Errorf("expected error message to contain 'upsert contact', got %v", err)
		}
	})
}

func TestContactStorage_GetByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	storage := postgres.NewContactStorage(db)
	ctx := context.Background()

	t.Run("found contact", func(t *testing.T) {
		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			return &mockRows{
				cols: []string{"id", "company_id", "telegram_user_id", "name", "username"},
				vals: [][]driver.Value{{int64(7), int64(1), int64(12345), "Bob", "bob123"}},
			}, nil
		}

		c, err := storage.GetByID(ctx, 7)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID != 7 || c.Name != "Bob" || c.Username != "bob123" {
			t.Errorf("unexpected contact: %#v", c)
		}
	})

	t.Run("not found translates to domain.ErrNotFound", func(t *testing.T) {
		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			return &mockRows{
				cols: []string{"id", "company_id", "telegram_user_id", "name", "username"},
				vals: [][]driver.Value{}, // empty -> sql.ErrNoRows in GetContext
			}, nil
		}

		_, err := storage.GetByID(ctx, 999)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected domain.ErrNotFound, got: %v", err)
		}
	})
}

func TestContactStorage_GetByTelegramID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	storage := postgres.NewContactStorage(db)
	ctx := context.Background()

	t.Run("found contact", func(t *testing.T) {
		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			return &mockRows{
				cols: []string{"id", "company_id", "telegram_user_id", "name", "username"},
				vals: [][]driver.Value{{int64(15), int64(2), int64(54321), "Carol", "carol_tg"}},
			}, nil
		}

		c, err := storage.GetByTelegramID(ctx, 2, 54321)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID != 15 || c.Name != "Carol" {
			t.Errorf("unexpected contact: %#v", c)
		}
	})

	t.Run("not found translates to domain.ErrNotFound", func(t *testing.T) {
		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			return &mockRows{
				cols: []string{"id", "company_id", "telegram_user_id", "name", "username"},
				vals: [][]driver.Value{},
			}, nil
		}

		_, err := storage.GetByTelegramID(ctx, 2, 99999)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected domain.ErrNotFound, got: %v", err)
		}
	})
}

func TestContactStorage_ListByCompany(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	storage := postgres.NewContactStorage(db)
	ctx := context.Background()

	t.Run("returns contacts slice", func(t *testing.T) {
		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			return &mockRows{
				cols: []string{"id", "company_id", "telegram_user_id", "name", "username"},
				vals: [][]driver.Value{
					{int64(1), int64(1), int64(100), "User 1", "u1"},
					{int64(2), int64(1), int64(200), "User 2", "u2"},
				},
			}, nil
		}

		list, err := storage.ListByCompany(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("expected 2 contacts, got %d", len(list))
		}
	})

	t.Run("empty list returns non-nil empty slice", func(t *testing.T) {
		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			return &mockRows{
				cols: []string{"id", "company_id", "telegram_user_id", "name", "username"},
				vals: [][]driver.Value{},
			}, nil
		}

		list, err := storage.ListByCompany(ctx, 99)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if list == nil {
			t.Fatal("expected non-nil slice, got nil")
		}
		if len(list) != 0 {
			t.Errorf("expected 0 contacts, got %d", len(list))
		}
	})
}

// ==========================================
// 2. ChatStorage Empirical Tests
// ==========================================

func TestChatStorage_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	storage := postgres.NewChatStorage(db)
	ctx := context.Background()

	t.Run("default fallbacks applied when fields are empty/zero", func(t *testing.T) {
		var capturedArgs []driver.Value

		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			capturedArgs = args
			return &mockRows{
				cols: []string{"id"},
				vals: [][]driver.Value{{int64(101)}},
			}, nil
		}

		// Chat with zero optional values
		chat := domain.Chat{
			AccountID: 5,
			ContactID: 8,
		}

		id, err := storage.Create(ctx, chat)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != 101 {
			t.Errorf("expected id 101, got %d", id)
		}

		if len(capturedArgs) != 8 {
			t.Fatalf("expected 8 args, got %d", len(capturedArgs))
		}

		// $1: account_id, $2: contact_id
		if capturedArgs[0] != int64(5) || capturedArgs[1] != int64(8) {
			t.Errorf("unexpected account/contact args: %#v", capturedArgs[:2])
		}
		// $3: status fallback -> domain.ChatStatusNewLead
		if str, ok := capturedArgs[2].(string); !ok || str != string(domain.ChatStatusNewLead) {
			t.Errorf("expected default status %v, got %v (%T)", domain.ChatStatusNewLead, capturedArgs[2], capturedArgs[2])
		}
		// $4: status_source fallback -> domain.StatusSourceAuto
		if str, ok := capturedArgs[3].(string); !ok || str != string(domain.StatusSourceAuto) {
			t.Errorf("expected default source %v, got %v (%T)", domain.StatusSourceAuto, capturedArgs[3], capturedArgs[3])
		}
		// $5: priority fallback -> domain.ChatPriorityNormal
		if str, ok := capturedArgs[4].(string); !ok || str != string(domain.ChatPriorityNormal) {
			t.Errorf("expected default priority %v, got %v (%T)", domain.ChatPriorityNormal, capturedArgs[4], capturedArgs[4])
		}
		// $6: lastMessageAt fallback -> non-zero time
		lastMsgTime, ok := capturedArgs[5].(time.Time)
		if !ok || lastMsgTime.IsZero() {
			t.Errorf("expected non-zero lastMessageAt time, got %v", capturedArgs[5])
		}
		// $7: awaiting_reply_since -> nil
		if capturedArgs[6] != nil {
			t.Errorf("expected nil awaiting_reply_since, got %v", capturedArgs[6])
		}
		// $8: unread_count -> 0
		if capturedArgs[7] != int64(0) && capturedArgs[7] != 0 {
			t.Errorf("expected 0 unread_count, got %v", capturedArgs[7])
		}
	})

	t.Run("explicit values preserved", func(t *testing.T) {
		var capturedArgs []driver.Value

		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			capturedArgs = args
			return &mockRows{
				cols: []string{"id"},
				vals: [][]driver.Value{{int64(102)}},
			}, nil
		}

		awaitingTime := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
		lastTime := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
		chat := domain.Chat{
			AccountID:          1,
			ContactID:          2,
			Status:             domain.ChatStatusNegotiation,
			StatusSource:       domain.StatusSourceManual,
			Priority:           domain.ChatPriorityHigh,
			LastMessageAt:      lastTime,
			AwaitingReplySince: &awaitingTime,
			UnreadCount:        3,
		}

		id, err := storage.Create(ctx, chat)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != 102 {
			t.Errorf("expected id 102, got %d", id)
		}

		if str, ok := capturedArgs[2].(string); !ok || str != string(domain.ChatStatusNegotiation) {
			t.Errorf("expected status %v, got %v", domain.ChatStatusNegotiation, capturedArgs[2])
		}
		if str, ok := capturedArgs[3].(string); !ok || str != string(domain.StatusSourceManual) {
			t.Errorf("expected source %v, got %v", domain.StatusSourceManual, capturedArgs[3])
		}
		if str, ok := capturedArgs[4].(string); !ok || str != string(domain.ChatPriorityHigh) {
			t.Errorf("expected priority %v, got %v", domain.ChatPriorityHigh, capturedArgs[4])
		}
		if capturedArgs[5] != lastTime {
			t.Errorf("expected lastMessageAt %v, got %v", lastTime, capturedArgs[5])
		}
		if tVal, ok := capturedArgs[6].(time.Time); !ok || !tVal.Equal(awaitingTime) {
			t.Errorf("expected awaitingReplySince %v, got %v (%T)", awaitingTime, capturedArgs[6], capturedArgs[6])
		}
		if capturedArgs[7] != int64(3) && capturedArgs[7] != 3 {
			t.Errorf("expected unreadCount 3, got %v", capturedArgs[7])
		}
	})
}

func TestChatStorage_GetByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	storage := postgres.NewChatStorage(db)
	ctx := context.Background()

	t.Run("found chat with non-nil awaiting_reply_since", func(t *testing.T) {
		fixedTime := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			return &mockRows{
				cols: []string{
					"id", "account_id", "contact_id", "status", "status_source", "priority",
					"last_message_at", "awaiting_reply_since", "unread_count",
				},
				vals: [][]driver.Value{
					{int64(20), int64(1), int64(2), "interested", "auto", "high", fixedTime, fixedTime, 5},
				},
			}, nil
		}

		chat, err := storage.GetByID(ctx, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if chat.ID != 20 || chat.Status != domain.ChatStatusInterested || chat.Priority != domain.ChatPriorityHigh || chat.UnreadCount != 5 {
			t.Errorf("unexpected chat fields: %#v", chat)
		}
		if chat.AwaitingReplySince == nil || *chat.AwaitingReplySince != fixedTime {
			t.Errorf("expected awaiting_reply_since %v, got %v", fixedTime, chat.AwaitingReplySince)
		}
	})

	t.Run("not found translates to domain.ErrNotFound", func(t *testing.T) {
		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			return &mockRows{
				cols: []string{
					"id", "account_id", "contact_id", "status", "status_source", "priority",
					"last_message_at", "awaiting_reply_since", "unread_count",
				},
				vals: [][]driver.Value{},
			}, nil
		}

		_, err := storage.GetByID(ctx, 404)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected domain.ErrNotFound, got %v", err)
		}
	})
}

func TestChatStorage_ListByCompany(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	storage := postgres.NewChatStorage(db)
	ctx := context.Background()

	t.Run("verifies query structure and joins", func(t *testing.T) {
		var capturedQuery string
		var capturedArgs []driver.Value

		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			capturedQuery = query
			capturedArgs = args
			return &mockRows{
				cols: []string{
					"id", "account_id", "contact_id", "status", "status_source", "priority",
					"last_message_at", "awaiting_reply_since", "unread_count",
				},
				vals: [][]driver.Value{},
			}, nil
		}

		status := domain.ChatStatusNegotiation
		empID := int64(3)
		_, err := storage.ListByCompany(ctx, 1, &status, &empID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cleanQ := strings.ToLower(strings.Join(strings.Fields(capturedQuery), " "))
		if !strings.Contains(cleanQ, "join accounts a on c.account_id = a.id") {
			t.Errorf("missing join with accounts: %s", capturedQuery)
		}
		if !strings.Contains(cleanQ, "where a.company_id = $1") {
			t.Errorf("missing where a.company_id = $1: %s", capturedQuery)
		}
		if !strings.Contains(cleanQ, "$2::varchar is null or c.status = $2") {
			t.Errorf("missing nullable $2::varchar check: %s", capturedQuery)
		}
		if !strings.Contains(cleanQ, "$3::bigint is null or a.employee_id = $3") {
			t.Errorf("missing nullable $3::bigint check: %s", capturedQuery)
		}
		if !strings.Contains(cleanQ, "order by c.last_message_at desc, c.id desc") {
			t.Errorf("missing order by c.last_message_at desc, c.id desc: %s", capturedQuery)
		}

		if len(capturedArgs) != 3 {
			t.Fatalf("expected 3 args, got %d", len(capturedArgs))
		}
		if capturedArgs[0] != int64(1) {
			t.Errorf("expected companyID 1, got %v", capturedArgs[0])
		}
		if str, ok := capturedArgs[1].(string); !ok || str != "negotiation" {
			t.Errorf("expected status arg 'negotiation', got %v", capturedArgs[1])
		}
		if emp, ok := capturedArgs[2].(int64); !ok || emp != 3 {
			t.Errorf("expected employeeID arg 3, got %v", capturedArgs[2])
		}
	})

	t.Run("handles nil status and nil employeeID cleanly", func(t *testing.T) {
		var capturedArgs []driver.Value

		mock.queryHandler = func(query string, args []driver.Value) (driver.Rows, error) {
			capturedArgs = args
			return &mockRows{
				cols: []string{
					"id", "account_id", "contact_id", "status", "status_source", "priority",
					"last_message_at", "awaiting_reply_since", "unread_count",
				},
				vals: [][]driver.Value{},
			}, nil
		}

		list, err := storage.ListByCompany(ctx, 1, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if list == nil {
			t.Fatal("expected non-nil empty slice, got nil")
		}

		if capturedArgs[1] != nil {
			t.Errorf("expected nil status arg, got %v", capturedArgs[1])
		}
		if capturedArgs[2] != nil {
			t.Errorf("expected nil employeeID arg, got %v", capturedArgs[2])
		}
	})
}

func TestChatStorage_UpdateStatus(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()
	storage := postgres.NewChatStorage(db)
	ctx := context.Background()

	t.Run("successful update with rows affected > 0", func(t *testing.T) {
		var capturedArgs []driver.Value

		mock.execHandler = func(query string, args []driver.Value) (driver.Result, error) {
			capturedArgs = args
			return mockResult{rowsAffected: 1}, nil
		}

		err := storage.UpdateStatus(ctx, 42, domain.ChatStatusUnhappy, domain.StatusSourceManual)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(capturedArgs) != 3 {
			t.Fatalf("expected 3 args, got %d", len(capturedArgs))
		}
		if capturedArgs[0] != string(domain.ChatStatusUnhappy) ||
			capturedArgs[1] != string(domain.StatusSourceManual) ||
			(capturedArgs[2] != int64(42) && capturedArgs[2] != 42) {
			t.Errorf("unexpected update args: %#v", capturedArgs)
		}
	})

	t.Run("rows affected == 0 translates to domain.ErrNotFound", func(t *testing.T) {
		mock.execHandler = func(query string, args []driver.Value) (driver.Result, error) {
			return mockResult{rowsAffected: 0}, nil
		}

		err := storage.UpdateStatus(ctx, 999, domain.ChatStatusLost, domain.StatusSourceManual)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected domain.ErrNotFound, got: %v", err)
		}
	})

	t.Run("db exec error is wrapped", func(t *testing.T) {
		mock.execHandler = func(query string, args []driver.Value) (driver.Result, error) {
			return nil, errors.New("foreign key / db error")
		}

		err := storage.UpdateStatus(ctx, 1, domain.ChatStatusLost, domain.StatusSourceManual)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "update chat status") {
			t.Errorf("expected error message to contain 'update chat status', got: %v", err)
		}
	})
}

// ==========================================
// 3. Schema & Migrations Empirical Parser Tests
// ==========================================

func TestMigrationsSchemaIntegrity(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	// Navigate to repo root
	repoRoot := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	upPath := filepath.Join(repoRoot, "migrations", "000003_create_contacts_and_chats.up.sql")
	downPath := filepath.Join(repoRoot, "migrations", "000003_create_contacts_and_chats.down.sql")

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

	t.Run("contacts table definition", func(t *testing.T) {
		// Must have table contacts
		if !regexp.MustCompile(`(?i)create\s+table\s+(if\s+not\s+exists\s+)?contacts`).MatchString(upContent) {
			t.Error("missing CREATE TABLE contacts")
		}
		// Columns
		reqCols := []string{
			`company_id\s+BIGINT\s+NOT\s+NULL\s+REFERENCES\s+companies\(id\)\s+ON\s+DELETE\s+CASCADE`,
			`telegram_user_id\s+BIGINT\s+NOT\s+NULL`,
			`name\s+VARCHAR\(255\)\s+NOT\s+NULL`,
			`username\s+VARCHAR\(255\)\s+NOT\s+NULL\s+DEFAULT\s+''`,
		}
		for _, col := range reqCols {
			if !regexp.MustCompile(`(?i)` + col).MatchString(upContent) {
				t.Errorf("contacts missing column or constraint matching pattern: %s", col)
			}
		}
		// Unique constraint
		if !regexp.MustCompile(`(?i)UNIQUE\s*\(\s*company_id\s*,\s*telegram_user_id\s*\)`).MatchString(upContent) {
			t.Error("missing UNIQUE (company_id, telegram_user_id) constraint")
		}
		// Index
		if !regexp.MustCompile(`(?i)create\s+index\s+(if\s+not\s+exists\s+)?idx_contacts_company_id\s+on\s+contacts\s*\(\s*company_id\s*\)`).MatchString(upContent) {
			t.Error("missing index idx_contacts_company_id on contacts(company_id)")
		}
	})

	t.Run("chats table definition", func(t *testing.T) {
		if !regexp.MustCompile(`(?i)create\s+table\s+(if\s+not\s+exists\s+)?chats`).MatchString(upContent) {
			t.Error("missing CREATE TABLE chats")
		}
		reqCols := []string{
			`account_id\s+BIGINT\s+NOT\s+NULL\s+REFERENCES\s+accounts\(id\)\s+ON\s+DELETE\s+CASCADE`,
			`contact_id\s+BIGINT\s+NOT\s+NULL\s+REFERENCES\s+contacts\(id\)\s+ON\s+DELETE\s+CASCADE`,
			`status\s+VARCHAR\(32\)\s+NOT\s+NULL\s+DEFAULT\s+'new_lead'`,
			`status_source\s+VARCHAR\(16\)\s+NOT\s+NULL\s+DEFAULT\s+'auto'\s+CHECK\s*\(\s*status_source\s+IN\s*\(\s*'auto'\s*,\s*'manual'\s*\)\s*\)`,
			`priority\s+VARCHAR\(16\)\s+NOT\s+NULL\s+DEFAULT\s+'normal'\s+CHECK\s*\(\s*priority\s+IN\s*\(\s*'low'\s*,\s*'normal'\s*,\s*'high'\s*\)\s*\)`,
			`last_message_at\s+TIMESTAMPTZ\s+NOT\s+NULL\s+DEFAULT\s+NOW\(\)`,
			`awaiting_reply_since\s+TIMESTAMPTZ`,
			`unread_count\s+INT\s+NOT\s+NULL\s+DEFAULT\s+0`,
		}
		for _, col := range reqCols {
			if !regexp.MustCompile(`(?i)` + col).MatchString(upContent) {
				t.Errorf("chats missing column or constraint matching pattern: %s", col)
			}
		}
		// Unique constraint
		if !regexp.MustCompile(`(?i)UNIQUE\s*\(\s*account_id\s*,\s*contact_id\s*\)`).MatchString(upContent) {
			t.Error("missing UNIQUE (account_id, contact_id) constraint")
		}
		// Required indexes on chats:
		indexes := []string{
			`idx_chats_account_id\s+on\s+chats\s*\(\s*account_id\s*\)`,
			`idx_chats_contact_id\s+on\s+chats\s*\(\s*contact_id\s*\)`,
			`idx_chats_status\s+on\s+chats\s*\(\s*status\s*\)`,
			`idx_chats_last_message_at\s+on\s+chats\s*\(\s*last_message_at\s+DESC\s*\)`,
		}
		for _, idx := range indexes {
			if !regexp.MustCompile(`(?i)create\s+index\s+(if\s+not\s+exists\s+)?` + idx).MatchString(upContent) {
				t.Errorf("missing index matching pattern: %s", idx)
			}
		}
	})

	t.Run("down migration drop order", func(t *testing.T) {
		chatsPos := strings.Index(strings.ToLower(downContent), "chats")
		contactsPos := strings.Index(strings.ToLower(downContent), "contacts")
		if chatsPos == -1 || contactsPos == -1 {
			t.Fatalf("down migration must drop both chats and contacts: %s", downContent)
		}
		if chatsPos > contactsPos {
			t.Errorf("chats must be dropped BEFORE contacts due to foreign key cascade: chatsPos=%d, contactsPos=%d", chatsPos, contactsPos)
		}
	})
}
