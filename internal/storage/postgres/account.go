package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kraconst/GO_server/internal/domain"
)

type AccountStorage struct {
	db *sqlx.DB
}

func NewAccountStorage(db *sqlx.DB) *AccountStorage {
	return &AccountStorage{db: db}
}

func (s *AccountStorage) Create(ctx context.Context, acc domain.Account) (int64, error) {
	query := `
		INSERT INTO accounts (company_id, employee_id, phone, username, telegram_user_id, status, session, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	var id int64
	createdAt := acc.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	err := s.db.QueryRowContext(ctx, query,
		acc.CompanyID,
		acc.EmployeeID,
		acc.Phone,
		acc.Username,
		acc.TelegramUserID,
		acc.Status,
		acc.Session,
		createdAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert account: %w", err)
	}

	// Записываем первоначальное назначение ответственного сотрудника
	if acc.EmployeeID > 0 {
		assignQuery := `
			INSERT INTO account_assignments (account_id, employee_id, from_time)
			VALUES ($1, $2, $3)
		`
		_, _ = s.db.ExecContext(ctx, assignQuery, id, acc.EmployeeID, createdAt)
	}

	return id, nil
}

func (s *AccountStorage) GetByID(ctx context.Context, id int64) (*domain.Account, error) {
	query := `
		SELECT a.id, a.company_id, a.employee_id, a.phone, a.username,
		       COALESCE(e.name, '') AS employee_name,
		       a.telegram_user_id, a.status, a.session, a.last_activity_at, a.created_at
		FROM accounts a
		LEFT JOIN employees e ON a.employee_id = e.id
		WHERE a.id = $1
	`
	var acc domain.Account
	err := s.db.GetContext(ctx, &acc, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select account by id: %w", err)
	}
	return &acc, nil
}

func (s *AccountStorage) ListByCompany(ctx context.Context, companyID int64) ([]domain.Account, error) {
	query := `
		SELECT a.id, a.company_id, a.employee_id, a.phone, a.username,
		       COALESCE(e.name, '') AS employee_name,
		       a.telegram_user_id, a.status, a.session, a.last_activity_at, a.created_at
		FROM accounts a
		LEFT JOIN employees e ON a.employee_id = e.id
		WHERE a.company_id = $1
		ORDER BY a.id ASC
	`
	var accounts []domain.Account
	err := s.db.SelectContext(ctx, &accounts, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("list accounts by company: %w", err)
	}
	return accounts, nil
}

func (s *AccountStorage) ListByEmployee(ctx context.Context, employeeID int64) ([]domain.Account, error) {
	query := `
		SELECT a.id, a.company_id, a.employee_id, a.phone, a.username,
		       COALESCE(e.name, '') AS employee_name,
		       a.telegram_user_id, a.status, a.session, a.last_activity_at, a.created_at
		FROM accounts a
		LEFT JOIN employees e ON a.employee_id = e.id
		WHERE a.employee_id = $1
		ORDER BY a.id ASC
	`
	var accounts []domain.Account
	err := s.db.SelectContext(ctx, &accounts, query, employeeID)
	if err != nil {
		return nil, fmt.Errorf("list accounts by employee: %w", err)
	}
	return accounts, nil
}

func (s *AccountStorage) UpdateStatus(ctx context.Context, id int64, status domain.AccountStatus) error {
	query := `UPDATE accounts SET status = $1 WHERE id = $2`
	res, err := s.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update account status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *AccountStorage) AssignEmployee(ctx context.Context, accountID, employeeID int64) error {
	now := time.Now()
	// Закрываем предыдущее активное назначение
	closeQuery := `
		UPDATE account_assignments
		SET to_time = $1
		WHERE account_id = $2 AND to_time IS NULL
	`
	_, err := s.db.ExecContext(ctx, closeQuery, now, accountID)
	if err != nil {
		return fmt.Errorf("close previous assignment: %w", err)
	}

	// Создаем новую запись назначения
	insertQuery := `
		INSERT INTO account_assignments (account_id, employee_id, from_time)
		VALUES ($1, $2, $3)
	`
	_, err = s.db.ExecContext(ctx, insertQuery, accountID, employeeID, now)
	if err != nil {
		return fmt.Errorf("insert new assignment: %w", err)
	}

	// Обновляем текущего ответственного в таблице accounts
	updateAccQuery := `UPDATE accounts SET employee_id = $1 WHERE id = $2`
	_, err = s.db.ExecContext(ctx, updateAccQuery, employeeID, accountID)
	if err != nil {
		return fmt.Errorf("update account employee_id: %w", err)
	}

	return nil
}
