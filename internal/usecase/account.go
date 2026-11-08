package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/kraconst/GO_server/internal/domain"
)

type AccountUsecase struct {
	accounts AccountRepository
}

func NewAccountUsecase(accounts AccountRepository) *AccountUsecase {
	return &AccountUsecase{
		accounts: accounts,
	}
}

type CreateAccountInput struct {
	CompanyID    int64
	EmployeeID   int64
	Phone        string
	Username     string
	EmployeeName string
}

func (u *AccountUsecase) Create(ctx context.Context, input CreateAccountInput) (*domain.Account, error) {
	phone := strings.TrimSpace(input.Phone)
	if phone == "" {
		return nil, fmt.Errorf("%w: phone is required", domain.ErrValidation)
	}

	acc := domain.Account{
		CompanyID:    input.CompanyID,
		EmployeeID:   input.EmployeeID,
		Phone:        phone,
		Username:     strings.TrimSpace(input.Username),
		EmployeeName: input.EmployeeName,
		Status:       domain.StatusActive,
	}

	id, err := u.accounts.Create(ctx, acc)
	if err != nil {
		return nil, fmt.Errorf("usecase create account: %w", err)
	}

	created, err := u.accounts.GetByID(ctx, id)
	if err != nil {
		acc.ID = id
		return &acc, nil
	}

	return created, nil
}

func (u *AccountUsecase) GetByID(ctx context.Context, id int64) (*domain.Account, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid id", domain.ErrValidation)
	}
	return u.accounts.GetByID(ctx, id)
}

func (u *AccountUsecase) List(ctx context.Context, companyID int64) ([]domain.Account, error) {
	return u.accounts.ListByCompany(ctx, companyID)
}

func (u *AccountUsecase) AssignEmployee(ctx context.Context, accountID, employeeID int64) error {
	if accountID <= 0 || employeeID <= 0 {
		return fmt.Errorf("%w: invalid account_id or employee_id", domain.ErrValidation)
	}
	return u.accounts.AssignEmployee(ctx, accountID, employeeID)
}
