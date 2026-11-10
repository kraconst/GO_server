package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kraconst/GO_server/internal/domain"
)

type ContactUsecase struct {
	contacts ContactRepository
}

func NewContactUsecase(contacts ContactRepository) *ContactUsecase {
	return &ContactUsecase{
		contacts: contacts,
	}
}

func (u *ContactUsecase) Create(ctx context.Context, c domain.Contact) (*domain.Contact, error) {
	if c.CompanyID <= 0 {
		return nil, fmt.Errorf("%w: invalid company_id", domain.ErrValidation)
	}
	if c.TelegramUserID <= 0 {
		return nil, fmt.Errorf("%w: invalid telegram_user_id", domain.ErrValidation)
	}
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	c.Username = strings.TrimSpace(c.Username)

	id, err := u.contacts.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("usecase create contact: %w", err)
	}

	created, err := u.contacts.GetByID(ctx, id)
	if err != nil {
		c.ID = id
		return &c, nil
	}
	return created, nil
}

func (u *ContactUsecase) FindOrCreateByTelegramID(ctx context.Context, companyID, tgUserID int64, name, username string) (*domain.Contact, error) {
	if companyID <= 0 {
		return nil, fmt.Errorf("%w: invalid company_id", domain.ErrValidation)
	}
	if tgUserID <= 0 {
		return nil, fmt.Errorf("%w: invalid telegram_user_id", domain.ErrValidation)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	username = strings.TrimSpace(username)

	contact, err := u.contacts.GetByTelegramID(ctx, companyID, tgUserID)
	if err == nil {
		return contact, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("usecase find contact: %w", err)
	}

	return u.Create(ctx, domain.Contact{
		CompanyID:      companyID,
		TelegramUserID: tgUserID,
		Name:           name,
		Username:       username,
	})
}

func (u *ContactUsecase) GetByID(ctx context.Context, id int64) (*domain.Contact, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid id", domain.ErrValidation)
	}
	return u.contacts.GetByID(ctx, id)
}

func (u *ContactUsecase) ListByCompany(ctx context.Context, companyID int64) ([]domain.Contact, error) {
	if companyID <= 0 {
		return nil, fmt.Errorf("%w: invalid company_id", domain.ErrValidation)
	}
	return u.contacts.ListByCompany(ctx, companyID)
}
