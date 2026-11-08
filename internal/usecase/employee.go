package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/kraconst/GO_server/internal/domain"
)

type EmployeeUsecase struct {
	employees EmployeeRepository
}

func NewEmployeeUsecase(employees EmployeeRepository) *EmployeeUsecase {
	return &EmployeeUsecase{
		employees: employees,
	}
}

type CreateEmployeeInput struct {
	CompanyID int64
	Name      string
	Position  string
}

func (u *EmployeeUsecase) Create(ctx context.Context, input CreateEmployeeInput) (*domain.Employee, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}

	emp := domain.Employee{
		CompanyID: input.CompanyID,
		Name:      name,
		Position:  strings.TrimSpace(input.Position),
		IsActive:  true,
	}

	id, err := u.employees.Create(ctx, emp)
	if err != nil {
		return nil, fmt.Errorf("usecase create employee: %w", err)
	}

	created, err := u.employees.GetByID(ctx, id)
	if err != nil {
		emp.ID = id
		return &emp, nil
	}
	return created, nil
}

func (u *EmployeeUsecase) GetByID(ctx context.Context, id int64) (*domain.Employee, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid id", domain.ErrValidation)
	}
	return u.employees.GetByID(ctx, id)
}

func (u *EmployeeUsecase) List(ctx context.Context, companyID int64) ([]domain.Employee, error) {
	return u.employees.ListByCompany(ctx, companyID)
}

func (u *EmployeeUsecase) Update(ctx context.Context, emp domain.Employee) error {
	if emp.ID <= 0 {
		return fmt.Errorf("%w: invalid id", domain.ErrValidation)
	}
	emp.Name = strings.TrimSpace(emp.Name)
	if emp.Name == "" {
		return fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	emp.Position = strings.TrimSpace(emp.Position)
	return u.employees.Update(ctx, emp)
}

func (u *EmployeeUsecase) Deactivate(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: invalid id", domain.ErrValidation)
	}
	return u.employees.Deactivate(ctx, id)
}
