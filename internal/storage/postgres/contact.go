package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kraconst/GO_server/internal/domain"
)

type ContactStorage struct {
	db *sqlx.DB
}

func NewContactStorage(db *sqlx.DB) *ContactStorage {
	return &ContactStorage{db: db}
}

// Create inserts a new contact or updates name/username on (company_id, telegram_user_id) conflict, returning the ID.
func (s *ContactStorage) Create(ctx context.Context, c domain.Contact) (int64, error) {
	query := `
		INSERT INTO contacts (company_id, telegram_user_id, name, username)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (company_id, telegram_user_id)
		DO UPDATE SET name = EXCLUDED.name, username = EXCLUDED.username
		RETURNING id
	`
	var id int64
	err := s.db.QueryRowContext(ctx, query,
		c.CompanyID,
		c.TelegramUserID,
		c.Name,
		c.Username,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert contact: %w", err)
	}
	return id, nil
}

// GetByID returns a contact by its primary key ID.
func (s *ContactStorage) GetByID(ctx context.Context, id int64) (*domain.Contact, error) {
	query := `
		SELECT id, company_id, telegram_user_id, name, username
		FROM contacts
		WHERE id = $1
	`
	var contact domain.Contact
	err := s.db.GetContext(ctx, &contact, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select contact by id: %w", err)
	}
	return &contact, nil
}

// GetByTelegramID returns a contact by company ID and Telegram user ID.
func (s *ContactStorage) GetByTelegramID(ctx context.Context, companyID, tgUserID int64) (*domain.Contact, error) {
	query := `
		SELECT id, company_id, telegram_user_id, name, username
		FROM contacts
		WHERE company_id = $1 AND telegram_user_id = $2
	`
	var contact domain.Contact
	err := s.db.GetContext(ctx, &contact, query, companyID, tgUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select contact by telegram id: %w", err)
	}
	return &contact, nil
}

// ListByCompany returns all contacts belonging to the given company.
func (s *ContactStorage) ListByCompany(ctx context.Context, companyID int64) ([]domain.Contact, error) {
	query := `
		SELECT id, company_id, telegram_user_id, name, username
		FROM contacts
		WHERE company_id = $1
		ORDER BY id ASC
	`
	contacts := make([]domain.Contact, 0)
	err := s.db.SelectContext(ctx, &contacts, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("list contacts by company: %w", err)
	}
	return contacts, nil
}
