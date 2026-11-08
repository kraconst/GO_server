package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kraconst/GO_server/internal/domain"
)

type EmployeeStorage struct {
	db *sqlx.DB
}

func NewEmployeeStorage(db *sqlx.DB) *EmployeeStorage {
	return &EmployeeStorage{db: db}
}

func (s *EmployeeStorage) Create(ctx context.Context, emp domain.Employee) (int64, error) {
	query := `
		INSERT INTO employees (company_id, name, position, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var id int64
	err := s.db.QueryRowContext(ctx, query,
		emp.CompanyID,
		emp.Name,
		emp.Position,
		emp.IsActive,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert employee: %w", err)
	}
	return id, nil
}

func (s *EmployeeStorage) GetByID(ctx context.Context, id int64) (*domain.Employee, error) {
	query := `
		SELECT id, company_id, name, position, is_active
		FROM employees
		WHERE id = $1
	`
	var emp domain.Employee
	err := s.db.GetContext(ctx, &emp, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select employee by id: %w", err)
	}
	return &emp, nil
}

func (s *EmployeeStorage) ListByCompany(ctx context.Context, companyID int64) ([]domain.Employee, error) {
	query := `
		SELECT id, company_id, name, position, is_active
		FROM employees
		WHERE company_id = $1
		ORDER BY id ASC
	`
	var employees []domain.Employee
	err := s.db.SelectContext(ctx, &employees, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("list employees by company: %w", err)
	}
	return employees, nil
}

func (s *EmployeeStorage) Update(ctx context.Context, emp domain.Employee) error {
	query := `
		UPDATE employees
		SET name = $1, position = $2
		WHERE id = $3 AND company_id = $4
	`
	res, err := s.db.ExecContext(ctx, query, emp.Name, emp.Position, emp.ID, emp.CompanyID)
	if err != nil {
		return fmt.Errorf("update employee: %w", err)
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

func (s *EmployeeStorage) Deactivate(ctx context.Context, id int64) error {
	query := `UPDATE employees SET is_active = FALSE WHERE id = $1`
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deactivate employee: %w", err)
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
